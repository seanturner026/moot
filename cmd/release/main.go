package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/github"
	util "github.com/seanturner026/serverless-release-dashboard/internal/util"
	"golang.org/x/oauth2"
)

// releaseEvent is an API Gateway POST which contains information necessary to create a release
type releaseEvent struct {
	GithubOwner    string `json:"github_owner"`
	GithubRepo     string `json:"github_repo"`
	BranchBase     string `json:"branch_base"`
	BranchHead     string `json:"branch_head"`
	ReleaseBody    string `json:"release_body"`
	ReleaseVersion string `json:"release_version"`
}

var (
	clientGithub *github.Client
	githubCtx    context.Context
)

// init authenicates with Github using the Github token provided environment variable
func init() {
	githubCtx = context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	tc := oauth2.NewClient(githubCtx, ts)
	clientGithub = github.NewClient(tc)
}

// createPullRequest generates a pull request on Github according to the ReleaseEvent
func createPullRequest(githubCtx context.Context, c *github.Client, e releaseEvent) (github.PullRequest, error) {
	pullRequestInfo := &github.NewPullRequest{
		Title: github.String(e.ReleaseVersion),
		Base:  github.String(e.BranchBase),
		Head:  github.String(e.BranchHead),
		Body:  github.String(e.ReleaseBody),
	}

	log.Printf("[INFO] creating %v pull request...", e.GithubRepo)
	resp, _, err := c.PullRequests.Create(
		githubCtx,
		e.GithubOwner,
		e.GithubRepo,
		pullRequestInfo,
	)

	if err != nil {
		log.Printf("[ERROR] unable to create %v pull request, %v", e.GithubRepo, err)
		return *resp, err
	}

	return *resp, nil
}

// mergePullRequest merges the pull request created by createPullRequest
func mergePullRequest(githubCtx context.Context, c *github.Client, prNumber int, e releaseEvent) (github.PullRequestMergeResult, error) {
	log.Printf("[INFO] merging pull request %v...", prNumber)
	mergeResult, _, err := c.PullRequests.Merge(
		githubCtx,
		e.GithubOwner,
		e.GithubRepo,
		prNumber,
		fmt.Sprintf("Merging pull request number %v", prNumber),
		&github.PullRequestOptions{},
	)

	if err != nil {
		log.Printf("[ERROR], unable to merge %v pull request %v, %v", e.GithubRepo, prNumber, err)
		return *mergeResult, err
	}
	return *mergeResult, nil
}

// createRelease creates a release on Github according to the ReleaseEvent
func createRelease(githubCtx context.Context, c *github.Client, e releaseEvent) error {
	releaseInfo := &github.RepositoryRelease{
		TargetCommitish: github.String(e.BranchBase),
		TagName:         github.String(e.ReleaseVersion),
		Name:            github.String(e.ReleaseVersion),
		Body:            github.String(e.ReleaseBody),
		Prerelease:      github.Bool(false),
	}

	log.Printf("[INFO] creating %v release version %v...", e.GithubRepo, e.ReleaseVersion)
	_, _, err := c.Repositories.CreateRelease(
		githubCtx,
		e.GithubOwner,
		e.GithubRepo,
		releaseInfo,
	)

	if err != nil {
		log.Printf("[ERROR] Unable to create %v release version %v, %v", e.GithubRepo, e.ReleaseVersion, err)
		return err
	}
	return nil
}

// handler executes the release and notification workflow
func handler(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := releaseEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	prResp, err := createPullRequest(githubCtx, clientGithub, e)
	if err != nil {
		message := fmt.Sprintf(
			"Could not create Github pull request for %v version %v, please check github for furhter details.",
			e.GithubRepo,
			e.ReleaseVersion,
		)
		resp := util.GenerateResponseBody(message, 404, err, headers)
		return resp, nil
	}

	mergeResp, err := mergePullRequest(githubCtx, clientGithub, *prResp.Number, e)
	if err != nil {
		message := fmt.Sprintf(
			"API request to merge github pull request %v for %v version %v failed, please check the pull request on github for further details.",
			*prResp.Number,
			e.GithubRepo,
			e.ReleaseVersion,
		)
		resp := util.GenerateResponseBody(message, 404, err, headers)
		return resp, nil
	}

	if !*mergeResp.Merged {
		log.Printf("[ERROR] %v pull request %v not merged", e.GithubRepo, *prResp.Number)
		message := fmt.Sprintf(
			"API request to merge github pull request %v for %v version %v failed, please check the pull request on github for further details.",
			*prResp.Number,
			e.GithubRepo,
			e.ReleaseVersion,
		)
		resp := util.GenerateResponseBody(message, 404, err, headers)
		return resp, nil
	}

	err = createRelease(githubCtx, clientGithub, e)
	if err != nil {
		message := fmt.Sprintf(
			"Unable to create %v release version %v on Github.",
			e.GithubRepo,
			e.ReleaseVersion,
		)
		resp := util.GenerateResponseBody(message, 404, err, headers)
		return resp, nil
	}

	util.PostToSlack(os.Getenv("WEBHOOK_URL"), fmt.Sprintf(
		"Starting release for %v version %v...\n\n%v",
		e.GithubRepo,
		e.ReleaseVersion,
		e.ReleaseBody,
	))

	resp := util.GenerateResponseBody(fmt.Sprintf("Released %v version %v successfully,", e.GithubRepo, e.ReleaseVersion), 200, nil, headers)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
