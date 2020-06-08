package cmd

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/bigkevmcd/image-hooks/pkg/client"
	"github.com/bigkevmcd/image-hooks/pkg/config"
	"github.com/bigkevmcd/image-hooks/pkg/updater"
)

func makeUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update a repository configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, _ := zap.NewProduction()
			defer func() {
				_ = logger.Sync() // flushes buffer, if any
			}()
			scmClient, err := factory.NewClient(viper.GetString("driver"), "", viper.GetString("github_token"))
			if err != nil {
				return fmt.Errorf("failed to create a git driver: %s", err)
			}

			sugar := logger.Sugar()
			updater := updater.New(sugar, client.New(scmClient), nil)
			return updater.UpdateRepository(context.Background(), configFromFlags(), viper.GetString("new-image-url"))
		},
	}

	cmd.Flags().String(
		"driver",
		"github",
		"go-scm driver name to use e.g. github, gitlab",
	)
	logIfError(viper.BindPFlag("driver", cmd.Flags().Lookup("driver")))

	cmd.Flags().String(
		"new-image-url",
		"",
		"Image URL to populate the file with e.g. myorg/my-image",
	)
	logIfError(viper.BindPFlag("new-image-url", cmd.Flags().Lookup("new-image-url")))
	logIfError(cmd.MarkFlagRequired("new-image-url"))

	addConfigFlags(cmd)

	return cmd
}

func addConfigFlags(cmd *cobra.Command) {
	cmd.Flags().String(
		"image-repo",
		"",
		"Image repo e.g. org/repo that is being updated - used in the created PR",
	)
	logIfError(viper.BindPFlag("image-repo", cmd.Flags().Lookup("image-repo")))
	logIfError(cmd.MarkFlagRequired("image-repo"))

	cmd.Flags().String(
		"source-repo",
		"",
		"Git repository to update e.g. org/repo",
	)
	logIfError(viper.BindPFlag("source-repo", cmd.Flags().Lookup("source-repo")))
	logIfError(cmd.MarkFlagRequired("source-repo"))

	cmd.Flags().String(
		"source-branch",
		"master",
		"Branch to fetch for updating",
	)
	logIfError(viper.BindPFlag("source-branch", cmd.Flags().Lookup("source-branch")))

	cmd.Flags().String(
		"file-path",
		"",
		"Path within the source-repo to update",
	)
	logIfError(viper.BindPFlag("file-path", cmd.Flags().Lookup("file-path")))
	logIfError(cmd.MarkFlagRequired("file-path"))

	cmd.Flags().String(
		"update-key",
		"",
		"JSON path within the file-path to update e.g. spec.template.spec.containers.0.image",
	)
	logIfError(viper.BindPFlag("update-key", cmd.Flags().Lookup("update-key")))
	logIfError(cmd.MarkFlagRequired("update-key"))

	cmd.Flags().String(
		"branch-generate-name",
		"",
		"Prefix for naming automatically generated branch, if empty, this will update source-branch",
	)
	logIfError(viper.BindPFlag("branch-generate-name", cmd.Flags().Lookup("branch-generate-name")))

	cmd.Flags().String(
		"github_token",
		"",
		"The GitHub token to authenticate requests",
	)
	logIfError(viper.BindPFlag("github_token", cmd.Flags().Lookup("github_token")))
}

func configFromFlags() *config.Repository {
	return &config.Repository{
		Name:               viper.GetString("image-repo"),
		SourceRepo:         viper.GetString("source-repo"),
		SourceBranch:       viper.GetString("source-branch"),
		FilePath:           viper.GetString("file-path"),
		UpdateKey:          viper.GetString("update-key"),
		BranchGenerateName: viper.GetString("branch-generate-name"),
	}
}
