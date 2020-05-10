package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/bigkevmcd/quay-imager/pkg/pushhook"
)

func makeHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "update repositories in response to quay.io hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, _ := zap.NewProduction()
			defer func() {
				err := logger.Sync() // flushes buffer, if any
				if err != nil {
					log.Fatal(err)
				}
			}()
			sugar := logger.Sugar()
			handler := pushhook.NewHandler(sugar)
			http.Handle("/pushhook", handler)
			listen := fmt.Sprintf(":%d", viper.GetInt("port"))
			return http.ListenAndServe(listen, nil)
		},
	}

	cmd.Flags().Int(
		"port",
		8080,
		"port to serve requests on",
	)
	logIfError(viper.BindPFlag("port", cmd.Flags().Lookup("port")))

	return cmd
}
