package handlers

import (
	"context"
	"fmt"

	"github.com/bigkevmcd/image-hooks/pkg/handlers/client"
	"github.com/bigkevmcd/image-hooks/pkg/handlers/config"
	"github.com/bigkevmcd/image-hooks/pkg/hooks"
	"github.com/bigkevmcd/image-hooks/pkg/syaml"
	"github.com/jenkins-x/go-scm/scm"
)

type logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

// New creates and returns a new Updater.
func New(l logger, c client.GitClient, cfgs *config.RepoConfiguration) *Updater {
	return &Updater{gitClient: c, configs: cfgs, nameGenerator: randomNameGenerator{timeSeed}, log: l}
}

// Updater can update a Git repo with an updated version of a file based on a
// RepositoryPushHook.
type Updater struct {
	configs       *config.RepoConfiguration
	gitClient     client.GitClient
	nameGenerator nameGenerator
	log           logger
}

func (u *Updater) Update(ctx context.Context, h hooks.PushEvent) error {
	cfg := u.configs.Find(h.EventRepository())
	if cfg == nil {
		u.log.Infow("failed to find repo", "name", h.EventRepository())
		return nil
	}
	u.log.Infow("found repo", "name", h.EventRepository(), "newURL", h.PushedImageURL())
	return u.UpdateRepository(ctx, cfg, h.EventRepository(), h.PushedImageURL())
}

func (u *Updater) UpdateRepository(ctx context.Context, cfg *config.Repository, repository, newURL string) error {
	current, err := u.gitClient.GetFile(ctx, cfg.SourceRepo, cfg.SourceBranch, cfg.FilePath)
	if err != nil {
		u.log.Errorw("failed to get file from repo", "error", err)
		return err
	}
	u.log.Infow("got existing file", "sha", current.Sha)

	u.log.Infow("new image reference", "image", newURL)
	updated, err := syaml.SetBytes(current.Data, cfg.UpdateKey, newURL)

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
	return u.createPRIfNecessary(ctx, cfg, newBranchName, repository)
}

func (u *Updater) createBranchIfNecessary(ctx context.Context, cfg *config.Repository, masterRef string) (string, error) {
	if cfg.BranchGenerateName == "" {
		u.log.Infow("no branchGenerateName configured, reusing source branch", "branch", cfg.SourceBranch)
		return cfg.SourceBranch, nil
	}

	newBranchName := u.nameGenerator.prefixedName(cfg.BranchGenerateName)
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
