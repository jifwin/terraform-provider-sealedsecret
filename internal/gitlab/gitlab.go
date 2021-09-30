package gitlab

import (
	"fmt"
	gl "github.com/xanzy/go-gitlab"
)

func CreateMergeRequest(url, token, sourceBranch, targetBranch string) error {
	git, err := gl.NewClient(token)
	if err != nil {
		return fmt.Errorf("unable to create new gitlab client: %w", err)
	}

	pid, err := getProjectId(url, git)
	if err != nil {
		return fmt.Errorf("unable to find project id for url %s: %w", url, err)
	}

	_, _, err = git.MergeRequests.CreateMergeRequest(pid, createMergeRequestOpts(targetBranch, sourceBranch))
	if err != nil {
		return fmt.Errorf("unable to create merge request: %w", err)
	}
	return nil
}

func getProjectId(url string, c *gl.Client) (int, error) {
	projects, _, err := c.Projects.ListProjects(&gl.ListProjectsOptions{})
	if err != nil {
		return 0, fmt.Errorf("unable to get projects: %w", err)
	}
	for _, project := range projects {
		if project.WebURL == url {
			return project.ID, nil
		}
	}
	return 0, fmt.Errorf("unable to find any project for url %s", url)
}

func createMergeRequestOpts(targetBranch, sourceBranch string) *gl.CreateMergeRequestOptions {
	var (
		title       = "Automated SealedSecret generation."
		description = "This MR was automatically created by the terraform-provider-sealedsecrets."
	)
	var (
		titlePtr        *string
		targetBranchPtr *string
		sourceBranchPtr *string
		descriptionPtr  *string
	)

	targetBranchPtr = &targetBranch
	sourceBranchPtr = &sourceBranch
	titlePtr = &title
	descriptionPtr = &description
	return &gl.CreateMergeRequestOptions{
		Title:        titlePtr,
		Description:  descriptionPtr,
		SourceBranch: sourceBranchPtr,
		TargetBranch: targetBranchPtr,
	}

}
