package main

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type gitlabController struct {
	MergeRequestSquash bool
	RemoveSourceBranch bool
	ProjectID          string
	Client             *gitlab.Client
}

func (app gitlabController) createMergeRequest(e releaseEvent) (gitlab.MergeRequest, error) {
	input := &gitlab.CreateMergeRequestOptions{
		Title:              gitlab.String(e.ReleaseVersion),
		Description:        gitlab.String(e.ReleaseBody),
		SourceBranch:       gitlab.String(e.BranchHead),
		TargetBranch:       gitlab.String(e.BranchBase),
		RemoveSourceBranch: gitlab.Bool(app.RemoveSourceBranch),
		Squash:             gitlab.Bool(false),
	}

	log.Info(fmt.Sprintf("creating %v merge request...", e.RepoName))
	resp, _, err := app.Client.MergeRequests.CreateMergeRequest(e.GitlabProjectID, input)
	if err != nil {
		log.Error(fmt.Sprintf("unable to create %v pull request, %v", e.RepoName, err))
		return *resp, err
	}

	return *resp, nil
}

func (app gitlabController) pollMergeRequestStatus(e releaseEvent, mergeRequestID int) error {
	input := &gitlab.GetMergeRequestsOptions{
		RenderHTML:                  gitlab.Bool(false),
		IncludeDivergedCommitsCount: gitlab.Bool(false),
		IncludeRebaseInProgress:     gitlab.Bool(false),
	}

	log.Info(fmt.Sprintf("checking %v merge request %v mergability...", e.RepoName, mergeRequestID))
	for i := 0; i < 7; i++ {
		resp, _, err := app.Client.MergeRequests.GetMergeRequest(e.GitlabProjectID, mergeRequestID, input)
		if err != nil {
			log.Error(fmt.Sprintf("unable to check %v merge request %v mergability, %v", e.RepoName, mergeRequestID, err))
			return err
		}
		if resp.MergeStatus == "can_be_merged" {
			return nil
		} else if resp.MergeStatus == "cannot_be_merged" {
			return errors.New("merge request has merge conflicts")
		}
	}

	return errors.New("merge request status never turned mergable")
}

func (app gitlabController) acceptMergeRequest(e releaseEvent, mergeRequestID int) error {
	input := &gitlab.AcceptMergeRequestOptions{
		MergeCommitMessage:       gitlab.String(fmt.Sprintf("Merging pull request number %v", mergeRequestID)),
		Squash:                   gitlab.Bool(false),
		ShouldRemoveSourceBranch: gitlab.Bool(true),
	}

	log.Info(fmt.Sprintf("completing %v merge request %v...", e.RepoName, mergeRequestID))
	_, _, err := app.Client.MergeRequests.AcceptMergeRequest(e.GitlabProjectID, mergeRequestID, input)
	if err != nil {
		log.Error(fmt.Sprintf("unable to merge %v merge request %v, %v", e.RepoName, mergeRequestID, err))
		return err
	}

	return nil
}

func (app gitlabController) createRelease(e releaseEvent) error {
	input := &gitlab.CreateReleaseOptions{
		Name:        gitlab.String(e.ReleaseVersion),
		TagName:     gitlab.String(e.ReleaseVersion),
		Description: gitlab.String(e.ReleaseBody),
		Ref:         gitlab.String(e.BranchBase),
	}

	log.Info(fmt.Sprintf("releasing %v version %v...", e.RepoName, e.ReleaseVersion))
	_, _, err := app.Client.Releases.CreateRelease(e.GitlabProjectID, input)
	if err != nil {
		log.Error(fmt.Sprintf("unable to create %v release %v, %v", e.RepoName, e.ReleaseVersion, err))
		return err
	}

	return nil
}

func (app application) releasesGitlabHandler(e releaseEvent, token string) (string, int) {
	clientGitlab, err := gitlab.NewClient(token)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	app.GL = gitlabController{
		ProjectID:          e.GitlabProjectID,
		MergeRequestSquash: false,
		RemoveSourceBranch: true,
		Client:             clientGitlab,
	}

	if !e.Hotfix {
		createMergeRequestResp, err := app.GL.createMergeRequest(e)
		if err != nil {
			message := fmt.Sprintf("Unable to create %v merge request, please check the merge request on gitlab for further details", e.RepoName)
			statusCode := 400
			return message, statusCode
		}

		err = app.GL.pollMergeRequestStatus(e, createMergeRequestResp.IID)
		if err != nil {
			message := fmt.Sprintf("Unable to merge %v merge request %v, please check the merge request on gitlab for further details",
				e.RepoName,
				createMergeRequestResp.IID)
			statusCode := 400
			return message, statusCode
		}

		err = app.GL.acceptMergeRequest(e, createMergeRequestResp.IID)
		if err != nil {
			message := fmt.Sprintf("Unable to complete %v merge request %v, please check the merge request on gitlab for further details",
				e.RepoName,
				createMergeRequestResp.IID)
			statusCode := 400
			return message, statusCode
		}
	}

	err = app.GL.createRelease(e)
	if err != nil {
		message := fmt.Sprintf("Unable to create %v release", e.RepoName)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Created %v release version %v on Gitlab.",
		e.RepoName,
		e.ReleaseVersion)
	statusCode := 200
	return message, statusCode
}
