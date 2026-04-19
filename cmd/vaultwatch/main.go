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

	ctx, stop := signal.NotifyContext(nil, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runner.Run(ctx); err != nil {
		log.Fatalf("runner: %v", err)
	}
}

func buildSenders(cfg *config.Config) []alert.Sender {
	var senders []alert.Sender
	senders = append(senders, sender.NewLogSender())

	if cfg.Alerting.Webhook.URL != "" {
		senders = append(senders, sender.NewWebhookSender(cfg.Alerting.Webhook.URL))
	}
	if cfg.Alerting.Slack.WebhookURL != "" {
		senders = append(senders, sender.NewSlackSender(cfg.Alerting.Slack.WebhookURL))
	}
	if cfg.Alerting.PagerDuty.RoutingKey != "" {
		senders = append(senders, sender.NewPagerDutySender(cfg.Alerting.PagerDuty.RoutingKey))
	}
	if cfg.Alerting.OpsGenie.APIKey != "" {
		senders = append(senders, sender.NewOpsGenieSender(cfg.Alerting.OpsGenie.APIKey))
	}
	if cfg.Alerting.VictorOps.URL != "" {
		senders = append(senders, sender.NewVictorOpsSender(cfg.Alerting.VictorOps.URL))
	}
	if cfg.Alerting.Datadog.APIKey != "" {
		senders = append(senders, sender.NewDatadogSender(cfg.Alerting.Datadog.APIKey))
	}
	if cfg.Alerting.SNS.TopicARN != "" {
		senders = append(senders, sender.NewSNSSender(cfg.Alerting.SNS.TopicARN))
	}
	if cfg.Alerting.Teams.WebhookURL != "" {
		senders = append(senders, sender.NewTeamsSender(cfg.Alerting.Teams.WebhookURL))
	}
	if cfg.Alerting.GoogleChat.WebhookURL != "" {
		senders = append(senders, sender.NewGoogleChatSender(cfg.Alerting.GoogleChat.WebhookURL))
	}
	if cfg.Alerting.Telegram.BotToken != "" && cfg.Alerting.Telegram.ChatID != "" {
		senders = append(senders, sender.NewTelegramSender(cfg.Alerting.Telegram.BotToken, cfg.Alerting.Telegram.ChatID))
	}
	if cfg.Alerting.Discord.WebhookURL != "" {
		senders = append(senders, sender.NewDiscordSender(cfg.Alerting.Discord.WebhookURL))
	}
	return senders
}
