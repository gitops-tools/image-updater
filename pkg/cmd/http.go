package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/gitops-tools/image-updater/pkg/applier"
	"github.com/gitops-tools/image-updater/pkg/config"
	"github.com/gitops-tools/image-updater/pkg/handler"
	"github.com/gitops-tools/image-updater/pkg/hooks"
	"github.com/gitops-tools/image-updater/pkg/hooks/docker"
	"github.com/gitops-tools/image-updater/pkg/hooks/quay"
	"github.com/gitops-tools/pkg/client"
)

func makeHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "update repositories in response to image hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			zapl, _ := zap.NewProduction()
			defer func() {
				_ = zapl.Sync() // flushes buffer, if any
			}()
			logger := zapr.NewLogger(zapl)
			scmClient, err := createClientFromViper()
			if err != nil {
				return fmt.Errorf("failed to create a git driver: %s", err)
			}
			f, err := os.Open(viper.GetString("config"))
			if err != nil {
				return err
			}
			defer f.Close()
			repos, err := config.Parse(f)
			if err != nil {
				return err
			}
			applier := applier.New(logger, client.New(scmClient), repos)
			p, err := parser()
			if err != nil {
				return err
			}
			handler := handler.New(logger, applier, p)
			http.Handle("/", handler)
			listen := fmt.Sprintf(":%d", viper.GetInt("port"))
			logger.Info("quay-hooks http starting", "port", viper.GetInt("port"), "parser", viper.GetString("parser"))
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
		"parser",
		"quay",
		"what driver to use to parse incoming webhooks e.g. quay, docker",
	)
	logIfError(viper.BindPFlag("parser", cmd.Flags().Lookup("parser")))

	cmd.Flags().String(
		"config",
		"/etc/image-updater/config.yaml",
		"repository configuration",
	)
	logIfError(viper.BindPFlag("config", cmd.Flags().Lookup("config")))

	return cmd
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
