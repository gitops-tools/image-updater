package handler

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/gitops-tools/image-updater/pkg/client/mock"
	"github.com/gitops-tools/image-updater/pkg/config"
	"github.com/gitops-tools/image-updater/pkg/hooks"
	"github.com/gitops-tools/image-updater/pkg/hooks/quay"
	"github.com/gitops-tools/image-updater/pkg/updater"
	"github.com/jenkins-x/go-scm/scm"
)

const (
	testQuayRepo   = "mynamespace/repository"
	testGitHubRepo = "testorg/testrepo"
	testFilePath   = "environments/test/services/service-a/test.yaml"
)

func TestHandler(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	updater := updater.New(logger.Sugar(), m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))
	h := New(logger.Sugar(), updater, quay.Parse)
	rec := httptest.NewRecorder()
	req := makeHookRequest(t, "testdata/push_hook.json")

	h.ServeHTTP(rec, req)

	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: "Image mynamespace/repository updated",
		Head:  "test-branch-a",
		Base:  "master",
		Body:  "Automated Image Update",
	})
}

func TestHandlerWithParseFailure(t *testing.T) {
	badParser := func(*http.Request) (hooks.PushEvent, error) {
		return nil, errors.New("failed")
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	m := mock.New(t)
	updater := updater.New(logger.Sugar(), m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))
	h := New(logger.Sugar(), updater, badParser)
	rec := httptest.NewRecorder()
	req := makeHookRequest(t, "testdata/push_hook.json")

	h.ServeHTTP(rec, req)

	m.RefutePullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: "Image mynamespace/repository updated",
		Head:  "test-branch-a",
		Base:  "master",
		Body:  "Automated Image Update",
	})
	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("StatusCode got %d, want %d", res.StatusCode, http.StatusInternalServerError)
	}
}

func TestHandlerWithFailureToUpdate(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	m := mock.New(t)
	updater := updater.New(logger.Sugar(), m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))
	h := New(logger.Sugar(), updater, quay.Parse)
	rec := httptest.NewRecorder()
	req := makeHookRequest(t, "testdata/push_hook.json")

	h.ServeHTTP(rec, req)

	m.RefutePullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Title: "Image mynamespace/repository updated",
		Head:  "test-branch-a",
		Base:  "master",
		Body:  "Automated Image Update",
	})
	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("StatusCode got %d, want %d", res.StatusCode, http.StatusInternalServerError)
	}
}

func makeHookRequest(t *testing.T, fixture string) *http.Request {
	t.Helper()
	b, err := ioutil.ReadFile(fixture)
	if err != nil {
		t.Fatalf("failed to read %s: %s", fixture, err)
	}
	req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
	req.Header.Add("Content-Type", "application/json")
	return req
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
