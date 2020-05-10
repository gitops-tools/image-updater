package hook

import (
	"context"
	"fmt"
	"testing"

	"github.com/bigkevmcd/quay-imager/pkg/hook/client/mock"
	"github.com/bigkevmcd/quay-imager/pkg/hook/config"
	"github.com/bigkevmcd/quay-imager/pkg/quay"
	"github.com/jenkins-x/go-scm/scm"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	testQuayRepo   = "testorg/testproject"
	testGitHubRepo = "testorg/testrepo"
	testFilePath   = "environments/test/services/service-a/test.yaml"
)

func TestUpdaterWithUnknownRepo(t *testing.T) {
	t.Skip()
}

func TestUpdaterWithKnownRepo(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)).Sugar()

	updater := New(logger, m, createConfigs())
	updater.nameGenerator = func(string) string {
		return "known-branch"
	}
	hook := createHook()

	updater.Update(context.Background(), hook)

	updated := m.GetUpdatedContents(testGitHubRepo, testFilePath, "known-branch")
	want := "test:\n  image: quay.io/testorg/repo\n"
	if s := string(updated); s != want {
		t.Fatalf("update failed, got %#v, want %#v", s, want)
	}
	m.AssertBranchCreated(testGitHubRepo, "known-branch", testSHA)
	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: fmt.Sprintf("Image %s updated", testQuayRepo),
		Body:  "Automated Image Update",
		Head:  "known-branch",
		Base:  "master",
	})
}

func createHook() *quay.RepositoryPushHook {
	return &quay.RepositoryPushHook{
		Repository: testQuayRepo,
		DockerURL:  "quay.io/testorg/repo",
	}
}

func createConfigs() *config.RepoConfiguration {
	return &config.RepoConfiguration{
		Repositories: []*config.Repository{
			{
				Name:               testQuayRepo,
				SourceRepo:         "testorg/testrepo",
				SourceBranch:       "master",
				FilePath:           testFilePath,
				UpdateKey:          "test.image",
				BranchGenerateName: "test-branch-",
			},
		},
	}
}
