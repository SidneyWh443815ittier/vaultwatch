package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
	"github.com/yourusername/vaultwatch/internal/config"
	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func main() {
	cfg, err := config.Load("vaultwatch.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	client, err := vault.NewClient(cfg)
	if err != nil {
		log.Fatalf("vault client: %v", err)
	}

	senders := buildSenders(cfg)
	notifier := alert.New(senders)
	leaseMon := monitor.New(cfg)
	runner := monitor.NewRunner(client, leaseMon, notifier, cfg)

	ctx, stop := signal.NotifyContext(os.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runner.Run(ctx); err != nil {
		log.Fatalf("runner: %v", err)
	}
}

func buildSenders(cfg *config.Config) []sender.Sender {
	var senders []sender.Sender

	senders = append(senders, sender.NewLogSender(log.Default()))

	if cfg.Alerting.Webhook.URL != "" {
		senders = append(senders, sender.NewWebhookSender(cfg.Alerting.Webhook.URL))
	}
	if cfg.Alerting.Slack.WebhookURL != "" {
		senders = append(senders, sender.NewSlackSender(cfg.Alerting.Slack.WebhookURL))
	}
	if cfg.Alerting.PagerDuty.IntegrationKey != "" {
		senders = append(senders, sender.NewPagerDutySender(cfg.Alerting.PagerDuty.IntegrationKey))
	}
	if cfg.Alerting.OpsGenie.APIKey != "" {
		senders = append(senders, sender.NewOpsGenieSender(cfg.Alerting.OpsGenie.APIKey))
	}
	if cfg.Alerting.GoogleChat.WebhookURL != "" {
		senders = append(senders, sender.NewGoogleChatSender(cfg.Alerting.GoogleChat.WebhookURL))
	}
	if cfg.Alerting.Email.Host != "" {
		senders = append(senders, sender.NewEmailSender(
			cfg.Alerting.Email.Host,
			cfg.Alerting.Email.Port,
			cfg.Alerting.Email.From,
			cfg.Alerting.Email.To,
		))
	}
	if cfg.Alerting.Discord.WebhookURL != "" {
		senders = append(senders, sender.NewDiscordSender(cfg.Alerting.Discord.WebhookURL))
	}
	if cfg.Alerting.Teams.WebhookURL != "" {
		senders = append(senders, sender.NewTeamsSender(cfg.Alerting.Teams.WebhookURL))
	}

	return senders
}
