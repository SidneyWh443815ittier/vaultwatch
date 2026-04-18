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
	cfgPath := "vaultwatch.yaml"
	if v := os.Getenv("VAULTWATCH_CONFIG"); v != "" {
		cfgPath = v
	}
	cfg, err := config.Load(cfgPath)
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
	if cfg.VictorOps.APIURL != "" {
		senders = append(senders, sender.NewVictorOpsSender(cfg.VictorOps.APIURL))
	}
	if cfg.Datadog.APIKey != "" {
		senders = append(senders, sender.NewDatadogSender(cfg.Datadog.APIKey))
	}
	if cfg.SNS.TopicARN != "" {
		senders = append(senders, sender.NewSNSSender(cfg.SNS.TopicARN))
	}
	if cfg.Teams.WebhookURL != "" {
		senders = append(senders, sender.NewTeamsSender(cfg.Teams.WebhookURL))
	}
	if cfg.Email.To != "" {
		senders = append(senders, sender.NewEmailSender(cfg.Email))
	}
	return senders
}
