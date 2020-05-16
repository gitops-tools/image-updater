package quay

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigkevmcd/image-hooks/pkg/hooks"
	"github.com/google/go-cmp/cmp"
)

var _ hooks.PushEvent = (*RepositoryPushHook)(nil)

func TestParseRepositoryPush(t *testing.T) {
	req := makeHookRequest(t, "testdata/push_hook.json")

	hook, err := ParseRepositoryPush(req)
	if err != nil {
		t.Fatal(err)
	}

	want := &RepositoryPushHook{
		Name:        "repository",
		Repository:  "mynamespace/repository",
		Namespace:   "mynamespace",
		DockerURL:   "quay.io/mynamespace/repository",
		Homepage:    "https://quay.io/repository/mynamespace/repository",
		UpdatedTags: []string{"latest"},
	}
	if diff := cmp.Diff(want, hook); diff != "" {
		t.Fatalf("hook doesn't match:\n%s", diff)
	}
}

func TestParseRepositoryPushWithNoBody(t *testing.T) {
	bodyErr := errors.New("just a test error")

	req := httptest.NewRequest("POST", "/", failingReader{err: bodyErr})

	_, err := ParseRepositoryPush(req)
	if err != bodyErr {
		t.Fatal("expected an error")
	}

}

func TestParseRepositoryPushWithUnparseableBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)

	_, err := ParseRepositoryPush(req)

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestPushedImageURL(t *testing.T) {
	hook := &RepositoryPushHook{
		Name:        "repository",
		DockerURL:   "quay.io/mynamespace/repository",
		UpdatedTags: []string{"latest"},
	}
	want := "quay.io/mynamespace/repository:latest"

	if u := hook.PushedImageURL(); u != want {
		t.Fatalf("got %s, want %s", u, want)
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

type failingReader struct {
	err error
}

func (f failingReader) Read(p []byte) (n int, err error) {
	return 0, f.err
}
func (f failingReader) Close() error {
	return f.err
}