package applier

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/gitops-tools/image-updater/pkg/config"
	"github.com/gitops-tools/image-updater/pkg/hooks"
	"github.com/gitops-tools/pkg/client"
	"github.com/gitops-tools/pkg/updater"
	"github.com/go-logr/logr"
)

var timeSeed = rand.New(rand.NewSource(time.Now().UnixNano()))

// New creates and returns a new Applier.
func New(l logr.Logger, c client.GitClient, cfgs *config.RepoConfiguration, opts ...updater.UpdaterFunc) *Applier {
	return &Applier{configs: cfgs, log: l, updater: updater.New(l, c, opts...)}
}

// Applier can update a Git repo with an updated version of a file based on a
// RepositoryPushHook.
type Applier struct {
	configs *config.RepoConfiguration
	log     logr.Logger
	updater *updater.Updater
}

// UpdateFromHook takes the incoming hook and triggers an update based on the
// configuration for the repo in the hook (if one matches).
func (u *Applier) UpdateFromHook(ctx context.Context, h hooks.PushEvent) error {
	cfg := u.configs.Find(h.EventRepository())
	if cfg == nil {
		u.log.Info("failed to find repo", "name", h.EventRepository())
		return nil
	}
	if cfg.TagMatch != "" {
		re, err := regexp.Compile(cfg.TagMatch)
		if err != nil {
			return fmt.Errorf("failed to compile TagMatch regular expression: %s", err)
		}
		if !re.MatchString(h.EventTag()) {
			u.log.Info("failed to match tag", "tag", h.EventTag(), "tagMatch", cfg.TagMatch)
			return nil
		}
	}
	u.log.Info("found repo", "name", h.EventRepository(), "newURL", h.PushedImageURL())
	return u.UpdateRepository(ctx, cfg, h.PushedImageURL())
}

// UpdateRepository does the job of fetching the existing file, updating it, and
// then optionally creating a PR.
func (u *Applier) UpdateRepository(ctx context.Context, cfg *config.Repository, newURL string) error {
	ci := updater.CommitInput{
		Repo:               cfg.SourceRepo,
		Filename:           cfg.FilePath,
		Branch:             cfg.SourceBranch,
		BranchGenerateName: cfg.BranchGenerateName,
		CommitMessage:      "Automatic update because an image was updated",
	}

	newBranch, err := u.updater.ApplyUpdateToFile(ctx, ci, updater.UpdateYAML(cfg.UpdateKey, newURL))
	if err != nil {
		u.log.Error(err, "failed to get file from repo")
		return err
	}
	u.log.Info("updated branch with image", "image", newURL, "branch", newBranch)

	// If we modified the original branch...
	if newBranch == cfg.SourceBranch {
		return nil
	}

	pullRequestInput := updater.PullRequestInput{
		Title:        "Automated image update",
		Body:         fmt.Sprintf("Automated update from %q", cfg.Name),
		Repo:         cfg.SourceRepo,
		NewBranch:    newBranch,
		SourceBranch: cfg.SourceBranch,
	}

	pr, err := u.updater.CreatePR(ctx, pullRequestInput)
	if err != nil {
		return fmt.Errorf("failed to create pull request in repo %s: %w", cfg.SourceRepo, err)
	}
	u.log.Info("created PullRequest", "link", pr.Link)
	return nil
}
