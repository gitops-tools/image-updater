package pubsubhandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/gitops-tools/pkg/client/mock"
	"github.com/gitops-tools/pkg/updater"
	"github.com/go-logr/zapr"
	"github.com/jenkins-x/go-scm/scm"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/gitops-tools/image-updater/pkg/applier"
	"github.com/gitops-tools/image-updater/pkg/config"
	"github.com/gitops-tools/image-updater/pkg/hooks/gcr"
)

const (
	testGcrRepo    = "gcr.io/mynamespace/repository"
	testGitHubRepo = "testorg/testrepo"
	testFilePath   = "environments/test/services/service-a/test.yaml"
)

func TestHandler(t *testing.T) {
	testSHA := "980a0d5f19a64b4b30a87d4206aade58726b60e3"
	logger := zapr.NewLogger(zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	m := mock.New(t)
	m.AddBranchHead(testGitHubRepo, "master", testSHA)
	m.AddFileContents(testGitHubRepo, testFilePath, "master", []byte("test:\n  image: old-image\n"))
	applier := applier.New(logger, m, createConfigs(), updater.NameGenerator(stubNameGenerator{"a"}))

	h := New(logger, applier, gcr.Parse)

	msg := makeHookMessage(t, "testdata/push_event.json")

	h.Handle(context.TODO(), msg)

	m.AssertPullRequestCreated(testGitHubRepo, &scm.PullRequestInput{
		Body:  fmt.Sprintf("Automated update from %q", testGcrRepo),
		Head:  "test-branch-a",
		Base:  "master",
		Title: "Automated image update",
	})
}

func makeHookMessage(t *testing.T, fixture string) *stubMessage {
	t.Helper()
	b, err := ioutil.ReadFile(fixture)
	if err != nil {
		t.Fatalf("failed to read %s: %s", fixture, err)
	}
	msg := &stubMessage{
		data: b,
	}
	return msg
}

func createConfigs() *config.RepoConfiguration {
	return &config.RepoConfiguration{
		Repositories: []*config.Repository{
			{
				Name:               testGcrRepo,
				SourceRepo:         testGitHubRepo,
				SourceBranch:       "master",
				FilePath:           testFilePath,
				UpdateKey:          "test.image",
				BranchGenerateName: "test-branch-",
			},
		},
	}
}

type stubMessage struct {
	data []byte
}

func (m *stubMessage) Ack()         {}
func (m *stubMessage) Data() []byte { return m.data }

type stubNameGenerator struct {
	name string
}

func (s stubNameGenerator) PrefixedName(p string) string {
	return p + s.name
}
