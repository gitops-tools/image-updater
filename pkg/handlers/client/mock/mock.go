package mock

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/bigkevmcd/image-hooks/pkg/hook/client"
	"github.com/jenkins-x/go-scm/scm"
)

var _ client.GitClient = (*MockClient)(nil)

// New creates and returns a new MockClient.
func New(t *testing.T) *MockClient {
	return &MockClient{
		t:                   t,
		files:               make(map[string][]byte),
		updatedFiles:        make(map[string][]byte),
		createdBranches:     make(map[string]bool),
		branchHeads:         make(map[string]string),
		createdPullRequests: make(map[string][]*scm.PullRequestInput),
	}
}

// MockClient implements the client.GitClient interface with an in-memory
// representation of files.
type MockClient struct {
	t                    *testing.T
	files                map[string][]byte
	GetFileErr           error
	updatedFiles         map[string][]byte
	UpdateFileErr        error
	createdBranches      map[string]bool
	CreateBranchErr      error
	branchHeads          map[string]string
	createdPullRequests  map[string][]*scm.PullRequestInput
	CreatePullRequestErr error
}

// GetFile implements the client.GitClient interface.
func (m *MockClient) GetFile(ctx context.Context, repo, ref, path string) (*scm.Content, error) {
	if m.GetFileErr != nil {
		return nil, m.GetFileErr
	}
	if b, ok := m.files[key(repo, path, ref)]; ok {
		return &scm.Content{Data: b, Sha: bytesSha1(b)}, nil
	}
	return nil, nil
}

// UpdateFile implements the client.GitClient interface.
func (m *MockClient) UpdateFile(ctx context.Context, repo, branch, path, message, previousSHA string, content []byte) error {
	if m.UpdateFileErr != nil {
		return m.UpdateFileErr
	}
	// TODO: Do we need something to validate the previousSHA?
	m.updatedFiles[key(repo, path, branch)] = content
	return nil
}

// CreatePullRequest implements the client.GitClient interface.
func (m *MockClient) CreatePullRequest(ctx context.Context, repo string, inp *scm.PullRequestInput) (*scm.PullRequest, error) {
	if m.CreatePullRequestErr != nil {
		return nil, m.CreatePullRequestErr
	}
	existing, ok := m.createdPullRequests[repo]
	if !ok {
		existing = []*scm.PullRequestInput{}
	}
	existing = append(existing, inp)
	m.createdPullRequests[repo] = existing
	return &scm.PullRequest{Number: len(existing) + 1}, nil
}

// CreateBranch implements the client.GitClient interface.
func (m *MockClient) CreateBranch(ctx context.Context, repo, branch, sha string) error {
	if m.CreateBranchErr != nil {
		return m.CreateBranchErr
	}
	m.createdBranches[key(repo, branch, sha)] = true
	return nil
}

// GetBranchHead implements the client.GitClient interface.
func (m *MockClient) GetBranchHead(ctx context.Context, repo, branch string) (string, error) {
	ref, ok := m.branchHeads[key(repo, branch)]
	if !ok {
		return "", errors.New("not found")
	}
	return ref, nil
}

// AddFileContents is a mock method for setting up a fixture for
// GetFileContents.
func (m *MockClient) AddFileContents(repo, path, ref string, body []byte) {
	m.files[key(repo, path, ref)] = body
}

// GetUpdatedContents returns the bytes captured by the mock implementation for
// UpdateFile.
func (m *MockClient) GetUpdatedContents(repo, path, ref string) []byte {
	c := m.updatedFiles[key(repo, path, ref)]
	return c
}

// AddBranchHead is a mock for setting up a response for GetBranchHead.
func (m *MockClient) AddBranchHead(repo, branch, sha string) {
	m.branchHeads[key(repo, branch)] = sha
}

// AssertBranchCreated fails if no matching branch was created using
// CreateBranch.
func (m *MockClient) AssertBranchCreated(repo, branch, sha string) {
	if _, ok := m.createdBranches[key(repo, branch, sha)]; !ok {
		m.t.Fatalf("branch %s not created in repo %s from sha %s", branch, repo, sha)
	}
}

// RefuteBranchCreated fails if a matching branch was created using
// CreateBranch.
func (m *MockClient) RefuteBranchCreated(repo, branch, sha string) {
	if _, ok := m.createdBranches[key(repo, branch, sha)]; ok {
		m.t.Fatalf("branch %s was created in repo %s from sha %s", branch, repo, sha)
	}
}

// AssertPullRequestCreated fails if no matching PullRequest was created.
func (m *MockClient) AssertPullRequestCreated(repo string, inp *scm.PullRequestInput) {
	for _, pr := range m.createdPullRequests[repo] {
		if reflect.DeepEqual(inp, pr) {
			return
		}
	}
	m.t.Fatalf("pullrequest not created in repo %s", repo)
}

// RefutePullRequestCreated fails if matching PullRequest was created.
func (m *MockClient) RefutePullRequestCreated(repo string, inp *scm.PullRequestInput) {
	for _, pr := range m.createdPullRequests[repo] {
		if reflect.DeepEqual(inp, pr) {
			m.t.Fatalf("pullrequest was created in repo %s", repo)
		}
	}
}

func key(s ...string) string {
	return strings.Join(s, ":")
}

func bytesSha1(b []byte) string {
	h := sha1.New()
	h.Write([]byte(b))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
