package client

import (
	"context"

	"github.com/jenkins-x/go-scm/scm"
)

// GitClient wraps go-scm's Client with a simplified API.
type GitClient interface {
	GetFile(ctx context.Context, repo, ref, path string) (*scm.Content, error)
	UpdateFile(ctx context.Context, repo, branch, path, message, previousSHA string, content []byte) error
	CreatePullRequest(ctx context.Context, repo string, inp *scm.PullRequestInput) (*scm.PullRequest, error)
	CreateBranch(ctx context.Context, repo, branch, sha string) error
	GetBranchHead(ctx context.Context, repo, branch string) (string, error)
}
