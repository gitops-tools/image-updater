package applier

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gitops-tools/image-updater/pkg/config"
	"github.com/gitops-tools/image-updater/pkg/hooks/quay"
	"github.com/gitops-tools/pkg/client/mock"
	"github.com/gitops-tools/pkg/updater"
	"github.com/go-logr/zapr"
	"github.com/jenkins-x/go-scm/scm"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	testQuayRepo   = "mynamespace/repository"
	testGitHubRepo = "testorg/testrepo"
	testFilePath   = "environments/test/services/service-a/test.yaml"
)

func TestUpdaterWithUnknownRepo(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	applier := makeApplier(t, m, createConfigs())
	hook := createHook()
	hook.Repository = "unknown/repo"

	err := applier.UpdateFromHook(context.Background(), hook)

	// A non-matching repo is not considered an error.
	if err != nil {
		t.Fatal(err)
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != "" {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.RefuteBranchCreated(testGitHubRepo, "test-branch-a", testSHA)

	m.RefutePullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: fmt.Sprintf("Image %s updated", testQuayRepo),
		Body:  "Automated Image Update",
		Head:  "test-branch-a",
		Base:  "master",
	})
}

func TestUpdaterWithNonMatchingTag(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	configs := createConfigs()
	configs.Repositories[0].BranchGenerateName = ""
	configs.Repositories[0].TagMatch = "^v.*"
	applier := makeApplier(t, m, configs)
	hook := createHook()

	err := applier.UpdateFromHook(context.Background(), hook)

	// A non-matching tag is not considered an error.
	if err != nil {
		t.Fatal(err)
	}

	m.AssertNoInteractions()
}

func TestUpdaterWithKnownRepo(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	applier := makeApplier(t, m, createConfigs())
	hook := createHook()

	err := applier.UpdateFromHook(context.Background(), hook)
	if err != nil {
		t.Fatal(err)
	}

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	want := "test:\n  image: quay.io/testorg/repo:production\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, want)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: "Automated image update",
		Body:  fmt.Sprintf("Automated update from %q", testQuayRepo),
		Head:  "test-branch-a",
		Base:  "master",
	})
}

// With no name-generator, the change should be made to master directly, rather
// than going through a PullRequest.
func TestUpdaterWithNoNameGenerator(t *testing.T) {
	sourceBranch := "production"
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, sourceBranch, []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, sourceBranch, testSHA)
	configs := createConfigs()
	configs.Repositories[0].BranchGenerateName = ""
	configs.Repositories[0].SourceBranch = sourceBranch
	applier := makeApplier(t, m, configs)
	hook := createHook()

	err := applier.UpdateFromHook(context.Background(), hook)
	if err != nil {
		t.Fatal(err)
	}

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, sourceBranch)
	want := "test:\n  image: quay.io/testorg/repo:production\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, want)
	}
	m.AssertNoBranchesCreated()
	m.AssertNoPullRequestsCreated()
}

func TestUpdaterWithMissingFile(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	applier := makeApplier(t, m, createConfigs())
	hook := createHook()
	testErr := errors.New("missing file")
	m.GetFileErr = testErr

	err := applier.UpdateFromHook(context.Background(), hook)

	if err != testErr {
		t.Fatalf("got %s, want %s", err, testErr)
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != "" {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.AssertNoBranchesCreated()
	m.AssertNoPullRequestsCreated()
}

func TestUpdaterWithBranchCreationFailure(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	applier := makeApplier(t, m, createConfigs())
	hook := createHook()
	testErr := errors.New("can't create branch")
	m.CreateBranchErr = testErr

	err := applier.UpdateFromHook(context.Background(), hook)

	if err.Error() != "failed to create branch: can't create branch" {
		t.Fatalf("got %s, want %s", err, "failed to create branch: can't create branch")
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != "" {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.AssertNoBranchesCreated()
	m.AssertNoPullRequestsCreated()
}

func TestUpdaterWithUpdateFileFailure(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	applier := makeApplier(t, m, createConfigs())
	hook := createHook()
	testErr := errors.New("can't update file")
	m.UpdateFileErr = testErr

	err := applier.UpdateFromHook(context.Background(), hook)

	if err.Error() != "failed to update file: can't update file" {
		t.Fatalf("got %s, want %s", err, "failed to update file: can't update file")
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	if s := string(updated); s != "" {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.RefutePullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: fmt.Sprintf("Image %s updated", testQuayRepo),
		Body:  "Automated Image Update",
		Head:  "test-branch-a",
		Base:  "master",
	})
}

func TestUpdaterWithCreatePullRequestFailure(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	applier := makeApplier(t, m, createConfigs())
	hook := createHook()
	testErr := errors.New("failure")
	m.CreatePullRequestErr = testErr

	err := applier.UpdateFromHook(context.Background(), hook)

	if err.Error() != "failed to create pull request in repo testorg/testrepo: failed to create a pull request: failure" {
		t.Fatalf("got %s, want %s", err, "failed to create a pull request: can't create pull-request")
	}
	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	want := "test:\n  image: quay.io/testorg/repo:production\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, "")
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.RefutePullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: fmt.Sprintf("Image %s updated", testQuayRepo),
		Body:  "Automated Image Update",
		Head:  "test-branch-a",
		Base:  "master",
	})
}

func TestUpdaterWithNonMasterSourceBranch(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "staging", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "staging", testSHA)
	configs := createConfigs()
	configs.Repositories[0].SourceBranch = "staging"
	applier := makeApplier(t, m, configs)
	hook := createHook()

	err := applier.UpdateFromHook(context.Background(), hook)
	if err != nil {
		t.Fatal(err)
	}

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "test-branch-a")
	want := "test:\n  image: quay.io/testorg/repo:production\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, want)
	}
	m.AssertBranchCreated(testGitHubRepo, "test-branch-a", testSHA)
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: "Automated image update",
		Body:  fmt.Sprintf("Automated update from %q", testQuayRepo),
		Head:  "test-branch-a",
		Base:  "staging",
	})
}

func makeApplier(t *testing.T, m *mock.MockClient, cfgs *config.RepoConfiguration) *Applier {
	logger := zapr.NewLogger(zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	applier := New(logger, m, cfgs, updater.NameGenerator(stubNameGenerator{name: "a"}))
	return applier
}

func createHook() *quay.RepositoryPushHook {
	return &quay.RepositoryPushHook{
		Repository:  testQuayRepo,
		DockerURL:   "quay.io/testorg/repo",
		UpdatedTags: []string{"production"},
	}
}

func createConfigs() *config.RepoConfiguration {
	return &config.RepoConfiguration{
		Repositories: []*config.Repository{
			{
				Name:               testQuayRepo,
				SourceRepo:         testGitHubRepo,
				SourceBranch:       "master",
				FilePath:           testFilePath,
				UpdateKey:          "test.image",
				BranchGenerateName: "test-branch-",
			},
		},
	}
}

type stubNameGenerator struct {
	name string
}

func (s stubNameGenerator) PrefixedName(p string) string {
	return p + s.name
}
