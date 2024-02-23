package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
	"github.com/spf13/cobra"

	"github.com/lmr-hh/easybell-billing-info/easybell"
)

var (
	client          *easybell.Client
	sendWebhook     bool
	teamsWebhookURL string
	teamsClient     *goteamsnotify.TeamsClient

	NationalQuota       time.Duration
	MobileQuota         time.Duration
	NationalMinutePrice float64
	MobileMinutePrice   float64
)

func init() {
	rootCommand.PersistentFlags().DurationVarP(&NationalQuota, "national-minutes", "n", 0, "The included monthly quota for national calls.")
	rootCommand.PersistentFlags().DurationVarP(&MobileQuota, "mobile-minutes", "m", 0, "The included monthly quota of mobile calls.")
	rootCommand.PersistentFlags().Float64Var(&NationalMinutePrice, "national-price", 0.0083, "The price per minute for national phone minutes over the quota.")
	rootCommand.PersistentFlags().Float64Var(&MobileMinutePrice, "mobile-price", 0.0824, "The price per minute for mobile phone minutes over the quota.")
	rootCommand.PersistentFlags().BoolVar(&sendWebhook, "teams-webhook", true, "Send the report to a teams webhook.")
	rootCommand.PersistentFlags().StringVarP(&teamsWebhookURL, "webhook-url", "u", "", "Teams Webhook URL to send notifications to.")
}

var rootCommand = &cobra.Command{
	Use:   "easybell-billing-info",
	Short: "Create easyBell usage reports.",
	Long: "Create automated usage reports from the call logs in your easyBell account and\n" +
		"send them to a Teams channel.",
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		client = easybell.NewClient()
		username := os.Getenv("EASYBELL_USERNAME")
		if username == "" {
			return errors.New("no username specified")
		}
		password := os.Getenv("EASYBELL_PASSWORD")
		if password == "" {
			return errors.New("no password specified")
		}
		if NationalQuota == 0 {
			var err error
			if NationalQuota, err = time.ParseDuration(os.Getenv("EASYBELL_NATIONAL_MINUTES")); err != nil {
				return fmt.Errorf("invalid national minutes: %w", err)
			}
		}
		if MobileQuota == 0 {
			var err error
			if MobileQuota, err = time.ParseDuration(os.Getenv("EASYBELL_MOBILE_MINUTES")); err != nil {
				return fmt.Errorf("invalid mobile minutes: %w", err)
			}
		}
		if teamsWebhookURL == "" {
			teamsWebhookURL = os.Getenv("EASYBELL_TEAMS_WEBHOOK")
		}
		if sendWebhook {
			teamsClient = goteamsnotify.NewTeamsClient()
			if err := teamsClient.ValidateWebhook(teamsWebhookURL); err != nil {
				return err
			}
		}
		return client.Login(username, password)
	},
}
