package gcr

import (
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/gitops-tools/image-updater/pkg/hooks"
)

var _ hooks.PushEvent = (*PushMessage)(nil)
var _ hooks.PushEventParser = Parse

func TestParse(t *testing.T) {
	event := readFixture(t, "testdata/push_event.json")

	hook, err := Parse(event)
	if err != nil {
		t.Fatal(err)
	}

	want := &PushMessage{
		Action: "INSERT",
		Digest: "gcr.io/mynamespace/repository@sha256:6ec128e26cd5",
		Tag:    "gcr.io/mynamespace/repository:latest",
	}
	if diff := cmp.Diff(want, hook); diff != "" {
		t.Fatalf("hook doesn't match:\n%s", diff)
	}
}

func TestPushedImageURL(t *testing.T) {
	hook := &PushMessage{
		Action: "INSERT",
		Digest: "gcr.io/mynamespace/repository@sha256:6ec128e26cd5",
		Tag:    "gcr.io/mynamespace/repository:latest",
	}
	want := "gcr.io/mynamespace/repository:latest"

	if u := hook.PushedImageURL(); u != want {
		t.Fatalf("got %s, want %s", u, want)
	}
}

func TestRepository(t *testing.T) {
	hook := &PushMessage{
		Action: "INSERT",
		Digest: "gcr.io/mynamespace/repository@sha256:6ec128e26cd5",
		Tag:    "gcr.io/mynamespace/repository:latest",
	}
	want := "gcr.io/mynamespace/repository"

	if u := hook.EventRepository(); u != want {
		t.Fatalf("got %s, want %s", u, want)
	}
}

func readFixture(t *testing.T, fixture string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(fixture)
	if err != nil {
		t.Fatalf("failed to read %s: %s", fixture, err)
	}
	return b
}
