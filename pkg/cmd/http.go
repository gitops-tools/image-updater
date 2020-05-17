package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/bigkevmcd/image-hooks/pkg/client"
	"github.com/bigkevmcd/image-hooks/pkg/config"
	"github.com/bigkevmcd/image-hooks/pkg/handler"
	"github.com/bigkevmcd/image-hooks/pkg/hooks"
	"github.com/bigkevmcd/image-hooks/pkg/hooks/docker"
	"github.com/bigkevmcd/image-hooks/pkg/hooks/quay"
	"github.com/bigkevmcd/image-hooks/pkg/updater"
)

func makeHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "update repositories in response to quay.io hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, _ := zap.NewProduction()
			defer func() {
				_ = logger.Sync() // flushes buffer, if any
			}()

			scmClient, err := factory.NewClient(viper.GetString("driver"), "", githubToken())
			if err != nil {
				return fmt.Errorf("failed to create a git driver: %s", err)
			}

			sugar := logger.Sugar()

			f, err := os.Open(viper.GetString("config"))
			if err != nil {
				return err
			}
			defer f.Close()
			repos, err := config.Parse(f)
			if err != nil {
				return err
			}
			updater := updater.New(sugar, client.New(scmClient), repos)
			p, err := parser()
			if err != nil {
				return err
			}
			handler := handler.New(sugar, updater, p)
			http.Handle("/", handler)
			listen := fmt.Sprintf(":%d", viper.GetInt("port"))
			sugar.Infow("quay-hooks http starting", "port", viper.GetInt("port"), "parser", viper.GetString("parser"))
			return http.ListenAndServe(listen, nil)
		},
	}

	cmd.Flags().Int(
		"port",
		8080,
		"port to serve requests on",
	)
	logIfError(viper.BindPFlag("port", cmd.Flags().Lookup("port")))

	cmd.Flags().String(
		"driver",
		"github",
		"go-scm driver name to use e.g. github, gitlab",
	)
	logIfError(viper.BindPFlag("driver", cmd.Flags().Lookup("driver")))

	cmd.Flags().String(
		"parser",
		"quay",
		"what driver to use to parse incoming webhooks e.g. quay, docker",
	)
	logIfError(viper.BindPFlag("parser", cmd.Flags().Lookup("parser")))

	cmd.Flags().String(
		"config",
		"/etc/image-hooks/config.yaml",
		"repository configuration",
	)
	logIfError(viper.BindPFlag("config", cmd.Flags().Lookup("config")))
	return cmd
}

func githubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func parser() (hooks.PushEventParser, error) {
	switch viper.GetString("parser") {
	case "quay":
		return quay.Parse, nil
	case "docker":
		return docker.Parse, nil
	default:
		return nil, fmt.Errorf("unknown parser: %s", viper.GetString("parser"))
	}
}
