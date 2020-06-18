package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initConfig)
}

func logIfError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func makeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "image-hooks",
		TraverseChildren: true,
		Short:            "Update YAML files in a Git service, with optional automated Pull Requests",
	}

	cmd.PersistentFlags().String(
		"driver",
		"github",
		"go-scm driver name to use e.g. github, gitlab",
	)
	logIfError(viper.BindPFlag("driver", cmd.PersistentFlags().Lookup("driver")))
	cmd.PersistentFlags().String(
		"github_token",
		"",
		"The GitHub token to authenticate requests",
	)
	logIfError(viper.BindPFlag("github_token", cmd.PersistentFlags().Lookup("github_token")))

	cmd.PersistentFlags().String(
		"api-endpoint",
		"",
		"The API endpoint to communicate with private GitLab/GitHub installations",
	)
	logIfError(viper.BindPFlag("api-endpoint", cmd.PersistentFlags().Lookup("api-endpoint")))

	cmd.AddCommand(makeHTTPCmd())
	cmd.AddCommand(makeUpdateCmd())
	return cmd
}

func initConfig() {
	viper.AutomaticEnv()
}

// Execute is the main entry point into this component.
func Execute() {
	if err := makeRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
