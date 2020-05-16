package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

var _ GitClient = (*SCMClient)(nil)

func TestGetFile(t *testing.T) {
	gock.New("https://api.github.com").
		Get("/repos/Codertocat/Hello-World/contents/config/my/file.yaml").
		MatchParam("ref", "master").
		Reply(http.StatusOK).
		Type("application/json").
		File("testdata/content.json")
	defer gock.Off()

	scmClient, err := factory.NewClient("github", "", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient)

	body, err := client.GetFile(context.TODO(), "Codertocat/Hello-World", "master", "config/my/file.yaml")
	if err != nil {
		t.Fatal(err)
	}
	want := mustParseJSONAsContent(t, "testdata/content.json")
	if diff := cmp.Diff(want, body); diff != "" {
		t.Fatalf("got a different body back: %s\n", diff)
	}
}

func TestGetFileWithErrorResponse(t *testing.T) {
	gock.New("https://api.github.com").
		Get("/repos/Codertocat/Hello-World/contents/config/my/file.yaml").
		MatchParam("ref", "master").
		Reply(http.StatusInternalServerError).
		BodyString("not found")
	defer gock.Off()

	scmClient, err := factory.NewClient("github", "", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient)

	_, err = client.GetFile(context.TODO(), "Codertocat/Hello-World", "master", "config/my/file.yaml")
	if err.Error() != "failed to get file config/my/file.yaml from repo Codertocat/Hello-World ref master: (500)" {
		t.Fatal(err)
	}
}

func TestUpdateFile(t *testing.T) {
	message := "just a test message"
	content := []byte("testing")
	branch := "my-test-branch"
	sha := "980a0d5f19a64b4b30a87d4206aade58726b60e3"

	encode := func(b []byte) string {
		return base64.StdEncoding.EncodeToString(b)
	}

	gock.New("https://api.github.com").
		Put("/repos/Codertocat/Hello-World/contents/config/my/file.yaml").
		MatchType("json").
		JSON(map[string]string{"message": message, "content": encode(content), "branch": branch, "sha": sha}).
		Reply(http.StatusCreated).
		Type("application/json").
		File("testdata/content.json")
	defer gock.Off()

	scmClient, err := factory.NewClient("github", "", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient)

	err = client.UpdateFile(context.TODO(), "Codertocat/Hello-World", branch, "config/my/file.yaml", message, "980a0d5f19a64b4b30a87d4206aade58726b60e3", []byte(`testing`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateBranch(t *testing.T) {
	sha := "aa218f56b14c9653891f9e74264a383fa43fefbd"

	gock.New("https://api.github.com").
		Post("/repos/Codertocat/Hello-World/git/refs").
		MatchType("json").
		JSON(map[string]string{"ref": "refs/heads/new-feature", "sha": sha}).
		Reply(http.StatusCreated).
		Type("application/json").
		File("testdata/content.json")
	defer gock.Off()

	scmClient, err := factory.NewClient("github", "", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient)

	err = client.CreateBranch(context.Background(), "Codertocat/Hello-World", "new-feature", sha)
	if err != nil {
		t.Fatal(err)
	}
	if !gock.IsDone() {
		t.Fatal("branch was not created")
	}
}

func TestCreatePullRequest(t *testing.T) {
	title := "Amazing new feature"
	body := "Please pull these awesome changes in!"
	head := "octocat:new-feature"
	base := "master"

	gock.New("https://api.github.com").
		Post("/repos/Codertocat/Hello-World/pulls").
		MatchType("json").
		JSON(map[string]string{"title": title, "body": body, "head": head, "base": base}).
		Reply(http.StatusCreated).
		Type("application/json").
		File("testdata/pr_create.json")
	defer gock.Off()

	input := &scm.PullRequestInput{
		Title: title,
		Body:  "Please pull these awesome changes in!",
		Head:  "octocat:new-feature",
		Base:  "master",
	}
	scmClient, err := factory.NewClient("github", "", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient)

	_, err = client.CreatePullRequest(context.Background(), "Codertocat/Hello-World", input)
	if err != nil {
		t.Fatal(err)
	}
	if !gock.IsDone() {
		t.Fatal("pull request was not created")
	}
}

func TestGetBranchHead(t *testing.T) {
	gock.New("https://api.github.com").
		Get("/repos/Codertocat/Hello-World/git/refs/heads/master").
		Reply(http.StatusOK).
		Type("application/json").
		File("testdata/single_ref.json")
	defer gock.Off()

	scmClient, err := factory.NewClient("github", "", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient)

	_, err = client.GetBranchHead(context.Background(), "Codertocat/Hello-World", "master")
	if err != nil {
		t.Fatal(err)
	}
	if !gock.IsDone() {
		t.Fatal("ref was not fetched")
	}
}

func mustParseJSONAsContent(t *testing.T, filename string) *scm.Content {
	t.Helper()
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatal(err)
	}
	content, err := base64.StdEncoding.DecodeString(data["content"].(string))
	if err != nil {
		t.Fatal(err)
	}
	return &scm.Content{
		Path: data["path"].(string),
		Sha:  data["sha"].(string),
		Data: content,
	}
}
