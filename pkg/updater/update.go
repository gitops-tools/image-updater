package updater

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gitops-tools/image-hooks/pkg/client"
	"github.com/gitops-tools/image-hooks/pkg/config"
	"github.com/gitops-tools/image-hooks/pkg/hooks"
	"github.com/gitops-tools/image-hooks/pkg/names"
	"github.com/gitops-tools/image-hooks/pkg/syaml"
	"github.com/jenkins-x/go-scm/scm"
)

var timeSeed = rand.New(rand.NewSource(time.Now().UnixNano()))

type logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

type updaterFunc func(u *Updater)

// NameGenerator is an option func for the Updater creation function.
func NameGenerator(g names.Generator) updaterFunc {
	return func(u *Updater) {
		u.nameGenerator = g
	}
}

// New creates and returns a new Updater.
func New(l logger, c client.GitClient, cfgs *config.RepoConfiguration, opts ...updaterFunc) *Updater {
	u := &Updater{gitClient: c, configs: cfgs, nameGenerator: names.New(timeSeed), log: l}
	for _, o := range opts {
		o(u)
	}
	return u
}

// Updater can update a Git repo with an updated version of a file based on a
// RepositoryPushHook.
type Updater struct {
	configs       *config.RepoConfiguration
	gitClient     client.GitClient
	nameGenerator names.Generator
	log           logger
}

// UpdateFromHook takes the incoming hook and triggers an update based on the
// configuration for the repo in the hook (if one matches).
func (u *Updater) UpdateFromHook(ctx context.Context, h hooks.PushEvent) error {
	cfg := u.configs.Find(h.EventRepository())
	if cfg == nil {
		u.log.Infow("failed to find repo", "name", h.EventRepository())
		return nil
	}
	u.log.Infow("found repo", "name", h.EventRepository(), "newURL", h.PushedImageURL())
	return u.UpdateRepository(ctx, cfg, h.PushedImageURL())
}

// UpdateRepository does the job of fetching the existing file, updating it, and
// then optionally creating a PR.
func (u *Updater) UpdateRepository(ctx context.Context, cfg *config.Repository, newURL string) error {
	current, err := u.gitClient.GetFile(ctx, cfg.SourceRepo, cfg.SourceBranch, cfg.FilePath)
	if err != nil {
		u.log.Errorw("failed to get file from repo", "error", err)
		return err
	}
	u.log.Infow("got existing file", "sha", current.Sha)
	u.log.Infow("new image reference", "image", newURL)

	updated, err := syaml.SetBytes(current.Data, cfg.UpdateKey, newURL)
	if err != nil {
		return err
	}
	masterRef, err := u.gitClient.GetBranchHead(ctx, cfg.SourceRepo, cfg.SourceBranch)
	if err != nil {
		return fmt.Errorf("failed to get branch head: %v", err)
	}
	newBranchName, err := u.createBranchIfNecessary(ctx, cfg, masterRef)
	if err != nil {
		return err
	}
	err = u.gitClient.UpdateFile(ctx, cfg.SourceRepo, newBranchName, cfg.FilePath, "Automatic update because an image was updated", current.Sha, updated)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}
	u.log.Infow("updated file", "filename", cfg.FilePath)
	return u.createPRIfNecessary(ctx, cfg, newBranchName, cfg.Name)
}

func (u *Updater) createBranchIfNecessary(ctx context.Context, cfg *config.Repository, masterRef string) (string, error) {
	if cfg.BranchGenerateName == "" {
		u.log.Infow("no branchGenerateName configured, reusing source branch", "branch", cfg.SourceBranch)
		return cfg.SourceBranch, nil
	}

	newBranchName := u.nameGenerator.PrefixedName(cfg.BranchGenerateName)
	u.log.Infow("generating new branch", "name", newBranchName)
	err := u.gitClient.CreateBranch(ctx, cfg.SourceRepo, newBranchName, masterRef)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}
	u.log.Infow("created branch", "branch", newBranchName, "ref", masterRef)
	return newBranchName, nil
}

func (u *Updater) createPRIfNecessary(ctx context.Context, cfg *config.Repository, newBranchName, repository string) error {
	if cfg.SourceBranch == newBranchName {
		return nil
	}
	pr, err := u.gitClient.CreatePullRequest(ctx, cfg.SourceRepo, &scm.PullRequestInput{
		Title: fmt.Sprintf("Image %s updated", repository),
		Body:  "Automated Image Update",
		Head:  newBranchName,
		Base:  "master",
	})
	if err != nil {
		return fmt.Errorf("failed to create a pull request: %w", err)
	}
	u.log.Infow("created PullRequest", "number", pr.Number)
	return nil
}
