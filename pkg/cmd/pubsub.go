package cmd

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/gitops-tools/image-updater/pkg/applier"
	"github.com/gitops-tools/image-updater/pkg/config"
	"github.com/gitops-tools/image-updater/pkg/hooks/gcr"
	"github.com/gitops-tools/image-updater/pkg/pubsubhandler"
	"github.com/gitops-tools/pkg/client"
	"github.com/go-logr/zapr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	projectIDFlag        = "project-id"
	subscriptionNameFlag = "subscription-name"
)

type message struct {
	data []byte
}

func (m *message) Ack()         {}
func (m *message) Data() []byte { return m.data }

func makePubsubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubsub",
		Short: "update repositories in response to gcr pubsub events",
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

			sub, err := createSubscriptionFromViper()
			if err != nil {
				return err
			}

			handler := pubsubhandler.New(logger, applier, gcr.Parse)

			return sub.Receive(context.Background(), func(ctx context.Context, msg *pubsub.Message) {
				handler.Handle(ctx, &message{
					data: msg.Data,
				})
			})
		},
	}

	cmd.Flags().String(
		"config",
		"/etc/image-updater/config.yaml",
		"repository configuration",
	)
	logIfError(viper.BindPFlag("config", cmd.Flags().Lookup("config")))

	cmd.Flags().String(
		projectIDFlag,
		"",
		"GCP project ID",
	)
	logIfError(viper.BindPFlag(projectIDFlag, cmd.Flags().Lookup(projectIDFlag)))
	logIfError(cmd.MarkFlagRequired(projectIDFlag))

	cmd.Flags().String(
		subscriptionNameFlag,
		"",
		"GCP subscription name",
	)
	logIfError(viper.BindPFlag(subscriptionNameFlag, cmd.Flags().Lookup(subscriptionNameFlag)))
	logIfError(cmd.MarkFlagRequired(subscriptionNameFlag))

	return cmd
}

func createSubscriptionFromViper() (*pubsub.Subscription, error) {
	ctx := context.Background()
	projectID := viper.GetString(projectIDFlag)
	subscriptionName := viper.GetString(subscriptionNameFlag)

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	sub := client.Subscription(subscriptionName)
	return sub, nil
}
