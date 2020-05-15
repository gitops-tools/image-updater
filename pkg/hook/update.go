package hook

import (
	"context"
	"fmt"

	"github.com/bigkevmcd/quay-imager/pkg/hook/client"
	"github.com/bigkevmcd/quay-imager/pkg/hook/config"
	"github.com/bigkevmcd/quay-imager/pkg/quay"
	"github.com/bigkevmcd/quay-imager/pkg/syaml"
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

func (u *Updater) Update(ctx context.Context, h *quay.RepositoryPushHook) error {
	cfg := u.configs.Find(h.Repository)
	if cfg == nil {
		u.log.Infow("failed to find repo", "name", h.Repository)
		return nil
	}
	u.log.Infow("found repo", "name", h.Repository)

	current, err := u.gitClient.GetFile(ctx, cfg.SourceRepo, cfg.SourceBranch, cfg.FilePath)
	if err != nil {
		u.log.Errorw("failed to get file from repo", "error", err)
		return err
	}
	u.log.Infow("got existing file", "sha", current.Sha)
	newURL := fmt.Sprintf("%s:%s", h.DockerURL, h.UpdatedTags[0])

	u.log.Infow("new image reference", "image", newURL)
	updated, err := syaml.SetBytes(current.Data, cfg.UpdateKey, newURL)

	masterRef, err := u.gitClient.GetBranchHead(ctx, cfg.SourceRepo, cfg.SourceBranch)
	newBranchName := u.nameGenerator.prefixedName(cfg.BranchGenerateName)
	err = u.gitClient.CreateBranch(ctx, cfg.SourceRepo, newBranchName, masterRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	u.log.Infow("created branch", "branch", newBranchName, "ref", masterRef)

	err = u.gitClient.UpdateFile(ctx, cfg.SourceRepo, newBranchName, cfg.FilePath, "Automatic update because an image was updated", current.Sha, updated)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}
	u.log.Infow("updated file", "filename", cfg.FilePath)

	pr, err := u.gitClient.CreatePullRequest(ctx, cfg.SourceRepo, &scm.PullRequestInput{
		Title: fmt.Sprintf("Image %s updated", h.Repository),
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
