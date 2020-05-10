package hook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRepositoryPush(t *testing.T) {
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
