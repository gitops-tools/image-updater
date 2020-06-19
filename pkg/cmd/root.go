package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	driverFlag      = "driver"
	apiEndpointFlag = "api-endpoint"
	authTokenFlag   = "auth_token"
	insecureFlag    = "insecure"
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
		driverFlag,
		"github",
		"go-scm driver name to use e.g. github, gitlab",
	)
	logIfError(viper.BindPFlag(driverFlag, cmd.PersistentFlags().Lookup(driverFlag)))
	cmd.PersistentFlags().String(
		authTokenFlag,
		"",
		"The token to authenticate requests to your Git service",
	)
	logIfError(viper.BindPFlag(authTokenFlag, cmd.PersistentFlags().Lookup(authTokenFlag)))

	cmd.PersistentFlags().String(
		apiEndpointFlag,
		"",
		"The API endpoint to communicate with private GitLab/GitHub installations",
	)
	logIfError(viper.BindPFlag(apiEndpointFlag, cmd.PersistentFlags().Lookup(apiEndpointFlag)))

	cmd.PersistentFlags().Bool(
		insecureFlag,
		false,
		"Allow insecure server connections when using SSL",
	)
	logIfError(viper.BindPFlag(insecureFlag, cmd.PersistentFlags().Lookup(insecureFlag)))

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
