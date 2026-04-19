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
	leaseMonitor := monitor.New(cfg)
	runner := monitor.NewRunner(client, leaseMonitor, notifier, cfg)

	ctx, stop := signal.NotifyContext(os.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runner.Run(ctx); err != nil {
		log.Fatalf("runner: %v", err)
	}
}

func buildSenders(cfg *config.Config) []alert.Sender {
	var senders []alert.Sender
	senders = append(senders, sender.NewLogSender())
	if cfg.Webhook.URL != "" {
		senders = append(senders, sender.NewWebhookSender(cfg.Webhook.URL))
	}
	if cfg.Slack.WebhookURL != "" {
		senders = append(senders, sender.NewSlackSender(cfg.Slack.WebhookURL))
	}
	if cfg.PagerDuty.RoutingKey != "" {
		senders = append(senders, sender.NewPagerDutySender(cfg.PagerDuty.RoutingKey))
	}
	if cfg.OpsGenie.APIKey != "" {
		senders = append(senders, sender.NewOpsGenieSender(cfg.OpsGenie.APIKey))
	}
	if cfg.VictorOps.URL != "" {
		senders = append(senders, sender.NewVictorOpsSender(cfg.VictorOps.URL))
	}
	if cfg.Datadog.APIKey != "" {
		senders = append(senders, sender.NewDatadogSender(cfg.Datadog.APIKey))
	}
	if cfg.Teams.WebhookURL != "" {
		senders = append(senders, sender.NewTeamsSender(cfg.Teams.WebhookURL))
	}
	if cfg.GoogleChat.WebhookURL != "" {
		senders = append(senders, sender.NewGoogleChatSender(cfg.GoogleChat.WebhookURL))
	}
	if cfg.Discord.WebhookURL != "" {
		senders = append(senders, sender.NewDiscordSender(cfg.Discord.WebhookURL))
	}
	if cfg.Splunk.URL != "" {
		senders = append(senders, sender.NewSplunkSender(cfg.Splunk.URL, cfg.Splunk.Token))
	}
	if cfg.Grafana.APIKey != "" {
		senders = append(senders, sender.NewGrafanaSender(cfg.Grafana.APIKey))
	}
	return senders
}
