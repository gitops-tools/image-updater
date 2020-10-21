package quay

import (
	"io/ioutil"
	"testing"

	"github.com/gitops-tools/image-updater/pkg/hooks"
	"github.com/google/go-cmp/cmp"
)

var _ hooks.PushEvent = (*RepositoryPushHook)(nil)
var _ hooks.PushEventParser = Parse

func TestParse(t *testing.T) {
	req := makeHookRequest(t, "testdata/push_hook.json")

	hook, err := Parse(req)
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

func TestEventRepository(t *testing.T) {
	hook := &RepositoryPushHook{
		Repository:  "mynamespace/repository",
		Name:        "repository",
		DockerURL:   "quay.io/mynamespace/repository",
		UpdatedTags: []string{"latest"},
	}
	want := "mynamespace/repository"

	if u := hook.EventRepository(); u != want {
		t.Fatalf("got %s, want %s", u, want)
	}
}

func TestEventTag(t *testing.T) {
	hook := &RepositoryPushHook{
		Repository:  "mynamespace/repository",
		Name:        "repository",
		DockerURL:   "quay.io/mynamespace/repository",
		UpdatedTags: []string{"v1", "latest"},
	}

	want := "v1"
	if u := hook.EventTag(); u != want {
		t.Fatalf("got %s, want %s", u, want)
	}
}

func makeHookRequest(t *testing.T, fixture string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(fixture)
	if err != nil {
		t.Fatalf("failed to read %s: %s", fixture, err)
	}
	return b
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
