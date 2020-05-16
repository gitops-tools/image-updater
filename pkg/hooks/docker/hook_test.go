package docker

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bigkevmcd/image-hooks/pkg/hooks"
)

var _ hooks.PushEvent = (*WebhookEvent)(nil)
var _ hooks.PushEventParser = ParseRequest

func TestParseRequest(t *testing.T) {
	req := makeHookRequest(t, "testdata/push_event.json")

	hook, err := ParseRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	want := &WebhookEvent{
		CallbackURL: "https://registry.hub.docker.com/u/svendowideit/testhook/hook/2141b5bi5i5b02bec211i4eeih0242eg11000a/",
		PushData: &PushData{
			Pusher: "trustedbuilder",
			Tag:    "latest",
			Images: []string{
				"27d47432a69bca5f2700e4dff7de0388ed65f9d3fb1ec645e2bc24c223dc1cc3",
				"51a9c7c1f8bb2fa19bcd09789a34e63f35abb80044bc10196e304f6634cc582c",
				"...",
			},
			PushedAt: 1.417566161e+09,
		},
		Repository: &Repository{
			RepoName:        "svendowideit/testhook",
			Name:            "testhook",
			Namespace:       "svendowideit",
			Owner:           "svendowideit",
			FullDescription: "Docker Hub based automated build from a GitHub repo",
			RepoURL:         "https://registry.hub.docker.com/u/svendowideit/testhook/",
			Dockerfile:      "#\n# BUILD\t\tdocker build -t svendowideit/apt-cacher .\n# RUN\t\tdocker run -d -p 3142:3142 -name apt-cacher-run apt-cacher\n#\n# and then you can run containers with:\n# \t\tdocker run -t -i -rm -e http_proxy http://192.168.1.2:3142/ debian bash\n#\nFROM\t\tubuntu\n\n\nVOLUME\t\t[/var/cache/apt-cacher-ng]\nRUN\t\tapt-get update ; apt-get install -yq apt-cacher-ng\n\nEXPOSE \t\t3142\nCMD\t\tchmod 777 /var/cache/apt-cacher-ng ; /etc/init.d/apt-cacher-ng start ; tail -f /var/log/apt-cacher-ng/*\n",
			Status:          "Active",
			IsPrivate:       true,
			IsTrusted:       true,
			DateCreated:     1.417494799e+09,
		},
	}
	if diff := cmp.Diff(want, hook); diff != "" {
		t.Fatalf("hook doesn't match:\n%s", diff)
	}
}

func TestParseRequestWithNoBody(t *testing.T) {
	bodyErr := errors.New("just a test error")

	req := httptest.NewRequest("POST", "/", failingReader{err: bodyErr})

	_, err := ParseRequest(req)
	if err != bodyErr {
		t.Fatal("expected an error")
	}

}

func TestParseRequestWithUnparseableBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)

	_, err := ParseRequest(req)

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestPushedImageURL(t *testing.T) {
	hook := &WebhookEvent{
		PushData: &PushData{
			Tag: "latest",
		},
		Repository: &Repository{
			RepoName: "mynamespace/repository",
		},
	}
	want := "mynamespace/repository:latest"

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
