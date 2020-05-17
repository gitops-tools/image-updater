package handlers

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/bigkevmcd/image-hooks/pkg/handlers/client/mock"
	"github.com/bigkevmcd/image-hooks/pkg/hooks"
	"github.com/bigkevmcd/image-hooks/pkg/hooks/quay"
	"github.com/jenkins-x/go-scm/scm"
)

func TestHandler(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
	m := mock.New(t)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	updater := New(logger.Sugar(), m, createConfigs())
	updater.nameGenerator = stubNameGenerator{"a"}
	h := NewHandler(logger.Sugar(), updater, quay.Parse)
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
	updater := New(logger.Sugar(), m, createConfigs())
	updater.nameGenerator = stubNameGenerator{"a"}
	h := NewHandler(logger.Sugar(), updater, badParser)
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
	updater := New(logger.Sugar(), m, createConfigs())
	updater.nameGenerator = stubNameGenerator{"a"}
	h := NewHandler(logger.Sugar(), updater, quay.Parse)
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
