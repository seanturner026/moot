package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// releaseEvent is an API Gateway POST which contains information necessary to create a release
type releaseEvent struct {
	GithubOwner    string `json:"github_owner"`
	GithubRepo     string `json:"github_repo"`
	BranchHead     string `json:"branch_head"`
	BranchBase     string `json:"branch_base"`
	ReleaseBody    string `json:"release_body"`
	ReleaseVersion string `json:"release_version"`
}

// response is of type APIGatewayProxyResponse which leverages the AWS Lambda Proxy Request
// functionality (default behavior)
type response events.APIGatewayProxyResponse

// slackRequestBody defines the schema for POSTs to Slack webhooks
type slackRequestBody struct {
	Text string `json:"text"`
}

var (
	clientGithub *github.Client
	githubCtx    context.Context
)

// init authenicates with Github using the Github token provided environment variable
func init() {
	log.Println("[INFO] authenticating with github token...")
	githubCtx = context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	tc := oauth2.NewClient(githubCtx, ts)
	clientGithub = github.NewClient(tc)
}

// createPullRequest generates a pull request on Github according to the ReleaseEvent. Any errors
// will send a slack message and exits 0.
func createPullRequest(githubCtx context.Context, c *github.Client, r releaseEvent) *github.PullRequest {
	pullRequestInfo := &github.NewPullRequest{
		Title: github.String(r.ReleaseVersion),
		Head:  github.String(r.BranchHead),
		Base:  github.String(r.BranchBase),
		Body:  github.String(r.ReleaseBody),
	}

	log.Printf("[INFO] creating %v pull request...", r.GithubRepo)
	pullRequestResponse, _, err := c.PullRequests.Create(
		githubCtx,
		r.GithubOwner,
		r.GithubRepo,
		pullRequestInfo,
	)

	if err != nil {
		log.Printf("[ERROR] unable to create %v pull request, %v", r.GithubRepo, err)
	}

	if pullRequestResponse.Mergeable != nil {
		log.Printf("[ERROR] Pull request %v not mergeable, %v", r.GithubRepo, err)
		postToSlack(fmt.Sprintf(
			"Github pull request for %v version %v is un-mergeable, please fix merge conflicts and re-release.",
			r.GithubRepo,
			r.ReleaseVersion,
		))
		os.Exit(0)
	}

	return pullRequestResponse
}

// mergePullRequest merges the pull request created by createPullRequest. Any errors will send a
// slack message and exits 0.
func mergePullRequest(githubCtx context.Context, c *github.Client, prNumber int, r releaseEvent) {
	log.Printf("[INFO] merging pull request %v...", prNumber)
	mergeResult, resp, err := c.PullRequests.Merge(
		githubCtx,
		r.GithubOwner,
		r.GithubRepo,
		prNumber,
		fmt.Sprintf("Merging pull request number %v", prNumber),
		&github.PullRequestOptions{},
	)

	if err != nil || resp.Response.StatusCode != 200 {
		log.Printf("[ERROR], unable to merge %v pull request %v, %v", r.GithubRepo, prNumber, err)
	}

	if !*mergeResult.Merged {
		log.Printf("[ERROR] %v pull request %v not merged", r.GithubRepo, prNumber)
		postToSlack(fmt.Sprintf(
			"API request to merge github pull request %v for %v version %v failed.",
			prNumber,
			r.GithubRepo,
			r.ReleaseVersion,
		))
		os.Exit(0)
	}
}

// createRelease creates a release on Github according to the ReleaseEvent. Any errors will send a
// slack message and exits 0
func createRelease(githubCtx context.Context, c *github.Client, r releaseEvent) {
	releaseInfo := &github.RepositoryRelease{
		TargetCommitish: github.String(r.BranchBase),
		TagName:         github.String(r.ReleaseVersion),
		Name:            github.String(r.ReleaseVersion),
		Body:            github.String(r.ReleaseVersion),
		Prerelease:      github.Bool(false),
	}

	log.Printf("[INFO] creating %v release version %v...", r.GithubRepo, r.ReleaseVersion)
	_, resp, err := c.Repositories.CreateRelease(
		githubCtx,
		r.GithubOwner,
		r.GithubRepo,
		releaseInfo,
	)

	if err != nil || resp.Response.StatusCode != 200 {
		log.Printf("[ERROR] Unable to create %v release version %v, %v", r.GithubRepo, r.ReleaseVersion, err)
		postToSlack(fmt.Sprintf(
			"Unable to create %v release version %v on Github.",
			r.GithubRepo,
			r.ReleaseVersion,
		))
		os.Exit(0)
	}
}

// postToSlack reads a webhookURL from the provided environment variable, and sends the message
// argument to the channel associated with the webhookURL.
func postToSlack(message string) {
	webhookURL := os.Getenv("WEBHOOK_URL")
	slackBody, _ := json.Marshal(slackRequestBody{Text: message})
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		log.Printf("[ERROR] Unable to marshal json, %v", err)
		os.Exit(0)
	}

	req.Header.Add("Content-Type", "application/json")
	clientSlack := &http.Client{Timeout: 10 * time.Second}
	resp, err := clientSlack.Do(req)
	if err != nil {
		log.Printf("[ERROR] Unable to form POST request, %v", err)
		os.Exit(0)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		log.Println("[ERROR] Non-ok response returned from Slack")
		os.Exit(0)
	}
}

// handler executes the release and notification workflow
func handler(ctx context.Context, r releaseEvent) (response, error) {
	pr := createPullRequest(ctx, clientGithub, r)
	mergePullRequest(ctx, clientGithub, *pr.Number, r)
	createRelease(ctx, clientGithub, r)
	postToSlack(fmt.Sprintf(
		"Starting release for %v version %v...",
		r.GithubRepo,
		r.ReleaseVersion,
	))

	body, err := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("Released %v version %v successfully,", r.GithubRepo, r.ReleaseVersion),
	})

	if err != nil {
		return response{StatusCode: 404}, err
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)

	resp := response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
