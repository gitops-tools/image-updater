package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gitops-tools/pkg/client/mock"
	"github.com/gitops-tools/pkg/updater"
	"github.com/go-logr/zapr"
	"github.com/jenkins-x/go-scm/scm"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/gitops-tools/image-updater/pkg/applier"
	"github.com/gitops-tools/image-updater/pkg/config"
	"github.com/gitops-tools/image-updater/pkg/hooks"
	"github.com/gitops-tools/image-updater/pkg/hooks/quay"
)

const (
	testQuayRepo   = "mynamespace/repository"
	testGitHubRepo = "testorg/testrepo"
	testFilePath   = "environments/test/services/service-a/test.yaml"
)

func TestHandler(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	logger := zapr.NewLogger(zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	m := mock.New(t)
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	h := New(logger, applier.New(logger, m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"})), quay.Parse)
	rec := httptest.NewRecorder()
	req := makeHookRequest(t, "testdata/push_hook.json")

	h.ServeHTTP(rec, req)

	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Body:  fmt.Sprintf("Automated update from %q", testQuayRepo),
		Head:  "test-branch-a",
		Base:  "master",
		Title: "Automated image update",
	})
}

func TestHandlerWithParseFailure(t *testing.T) {
	badParser := func(payload []byte) (hooks.PushEvent, error) {
		return nil, errors.New("failed")
	}
	logger := zapr.NewLogger(zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	m := mock.New(t)
	applier := applier.New(logger, m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))
	h := New(logger, applier, badParser)
	rec := httptest.NewRecorder()
	req := makeHookRequest(t, "testdata/push_hook.json")

	h.ServeHTTP(rec, req)

	m.AssertNoPullRequestsCreated()
	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("StatusCode got %d, want %d", res.StatusCode, http.StatusInternalServerError)
	}
}

func TestHandlerWithFailureToUpdate(t *testing.T) {
	logger := zapr.NewLogger(zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	m := mock.New(t)
	applier := applier.New(logger, m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))
	h := New(logger, applier, quay.Parse)
	rec := httptest.NewRecorder()
	req := makeHookRequest(t, "testdata/push_hook.json")

	h.ServeHTTP(rec, req)

	m.AssertNoPullRequestsCreated()
	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("StatusCode got %d, want %d", res.StatusCode, http.StatusInternalServerError)
	}
}

func TestParseWithNoBody(t *testing.T) {
	logger := zapr.NewLogger(zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	m := mock.New(t)
	applier := applier.New(logger, m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))
	h := New(logger, applier, quay.Parse)
	bodyErr := errors.New("just a test error")

	req := httptest.NewRequest("POST", "/", failingReader{err: bodyErr})

	_, err := h.parse(req)
	if err != bodyErr {
		t.Fatal("expected an error")
	}
}

func TestParseWithUnparseableBody(t *testing.T) {
	logger := zapr.NewLogger(zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	m := mock.New(t)
	applier := applier.New(logger, m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))
	h := New(logger, applier, quay.Parse)

	req := httptest.NewRequest("POST", "/", nil)

	_, err := h.parse(req)

	if err == nil {
		t.Fatal("expected an error")
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

type failingReader struct {
	err error
}

func (f failingReader) Read(p []byte) (n int, err error) {
	return 0, f.err
}
func (f failingReader) Close() error {
	return f.err
}
