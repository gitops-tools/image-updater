package client

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
)

// New creates and returns a new SCMClient.
func New(c *scm.Client) *SCMClient {
	return &SCMClient{scmClient: c}
}

// SCMClient is a wrapper for the go-scm scm.Client with a simplified API.
type SCMClient struct {
	scmClient *scm.Client
}

// GetFile reads the specific revision of a file from a repository.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMClient) GetFile(ctx context.Context, repo, ref, path string) (*scm.Content, error) {
	content, r, err := c.scmClient.Contents.Find(ctx, repo, path, ref)
	if r != nil && isErrorStatus(r.Status) {
		return nil, scmError{msg: fmt.Sprintf("failed to get file %s from repo %s ref %s", path, repo, ref), Status: r.Status}
	}
	if err != nil {
		return nil, err
	}
	return content, nil
}

// CreateBranch will create a new branch in the repo from the SHA.
func (c *SCMClient) CreateBranch(ctx context.Context, repo, branch, sha string) error {
	if isGitHub(c.scmClient) {
		branch = fmt.Sprintf("refs/heads/%s", branch)
	}
	_, _, err := c.scmClient.Git.CreateRef(ctx, repo, branch, sha)
	return err
}

// CreatePullRequest creates a PullRequest with the provided input.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMClient) CreatePullRequest(ctx context.Context, repo string, inp *scm.PullRequestInput) (*scm.PullRequest, error) {
	pr, _, err := c.scmClient.PullRequests.Create(ctx, repo, inp)
	return pr, err
}

// UpdateFile updates an existing file in a repository.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMClient) UpdateFile(ctx context.Context, repo, branch, path, message, previousSHA string, content []byte) error {
	params := scm.ContentParams{
		Message: message,
		Data:    content,
		Branch:  branch,
		Sha:     previousSHA,
	}
	r, err := c.scmClient.Contents.Update(ctx, repo, path, &params)
	if err != nil {
		return err
	}
	if isErrorStatus(r.Status) {
		return scmError{msg: fmt.Sprintf("failed to update file %s in repo %s branch %s", path, repo, branch), Status: r.Status}
	}
	return nil
}

// GetBranchHead gets the head SHA for a specific branch.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMClient) GetBranchHead(ctx context.Context, repo, branch string) (string, error) {
	sha, _, err := c.scmClient.Git.FindRef(ctx, repo, fmt.Sprintf("heads/%s", branch))
	return sha, err
}

func isGitHub(c *scm.Client) bool {
	return c.Driver == scm.DriverGithub
}

func isErrorStatus(i int) bool {
	return i >= 400
}
