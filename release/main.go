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

// ReleaseEvent is a POST sent by API Gateway which contains information necessary to create a
// release
type ReleaseEvent struct {
	GithubOwner    string `json:"github_owner"`
	GithubRepo     string `json:"github_repo"`
	BranchHead     string `json:"branch_head"`
	BranchBase     string `json:"branch_base"`
	ReleaseBody    string `json:"release_body"`
	ReleaseVersion string `json:"release_version"`
}

// Response is of type APIGatewayProxyResponse which leverages the AWS Lambda Proxy Request
// functionality (default behavior)
type Response events.APIGatewayProxyResponse

// SlackRequestBody defines the schema for POSTs to Slack webhooks
type SlackRequestBody struct {
	Text string `json:"text"`
}

var (
	client    *github.Client
	githubCtx context.Context
)

func init() {
	githubCtx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(githubCtx, ts)
	client = github.NewClient(tc)
}

func createPullRequest(githubCtx context.Context, c *github.Client, r ReleaseEvent) *github.PullRequest {
	pullRequestInfo := &github.NewPullRequest{
		Title: github.String(r.ReleaseVersion),
		Head:  github.String(r.BranchHead),
		Base:  github.String(r.BranchBase),
		Body:  github.String(r.ReleaseBody),
	}

	pullRequestResponse, _, err := c.PullRequests.Create(
		githubCtx,
		r.GithubOwner,
		r.GithubRepo,
		pullRequestInfo,
	)

	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	if pullRequestResponse.Mergeable != nil {
		log.Printf("[ERROR] Pull request not mergeable, %v", err)
		postToSlack(fmt.Sprintf(
			"Github pull request for %v version %v is un-mergeable, please fix merge conflicts and re-release.",
			r.GithubRepo,
			r.ReleaseVersion,
		))
		os.Exit(0)
	}

	return pullRequestResponse
}

func mergePullRequest(githubCtx context.Context, c *github.Client, prNumber int, r ReleaseEvent) {
	mergeResult, _, err := c.PullRequests.Merge(
		githubCtx,
		r.GithubOwner,
		r.GithubRepo,
		prNumber,
		fmt.Sprintf("Merging pull request number %v", prNumber),
		&github.PullRequestOptions{},
	)

	if err != nil {
		log.Println(err)
	}

	if !*mergeResult.Merged {
		log.Println("[ERROR] Pull request not merged")
		postToSlack(fmt.Sprintf(
			"API request to merge github pull request for %v version %v failed.",
			r.GithubRepo,
			r.ReleaseVersion,
		))
		os.Exit(0)
	}
}

func createRelease(githubCtx context.Context, c *github.Client, r ReleaseEvent) {
	releaseInfo := &github.RepositoryRelease{
		TargetCommitish: github.String(r.BranchBase),
		TagName:         github.String(r.ReleaseVersion),
		Name:            github.String(r.ReleaseVersion),
		Body:            github.String(r.ReleaseVersion),
		Prerelease:      github.Bool(false),
	}
	_, _, err := c.Repositories.CreateRelease(
		githubCtx,
		r.GithubOwner,
		r.GithubRepo,
		releaseInfo,
	)

	if err != nil {
		log.Printf("[ERROR] Unable to create release, %v", err)
		postToSlack(fmt.Sprintf(
			"Unable to create %v release version %v on Github.",
			r.GithubRepo,
			r.ReleaseVersion,
		))
		os.Exit(0)
	}
}

func postToSlack(message string) {
	webhookURL := os.Getenv("WEBHOOK_URL")
	slackBody, _ := json.Marshal(SlackRequestBody{Text: message})
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		log.Printf("[ERROR] Unable to marshal-ing json, %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Unable to form POST request, %v", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		log.Println("[ERROR] Non-ok response returned from Slack")
	}
}

// Handler executes the release and notification workflow
func Handler(ctx context.Context, r ReleaseEvent) (Response, error) {
	pr := createPullRequest(ctx, client, r)
	mergePullRequest(ctx, client, *pr.Number, r)
	createRelease(ctx, client, r)
	postToSlack(fmt.Sprintf(
		"Starting release for %v version %v...",
		r.GithubRepo,
		r.ReleaseVersion,
	))

	body, err := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("Released %v version %v successfully,", r.GithubRepo, r.ReleaseVersion),
	})

	if err != nil {
		return Response{StatusCode: 404}, err
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)

	resp := Response{
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
	lambda.Start(Handler)
}
