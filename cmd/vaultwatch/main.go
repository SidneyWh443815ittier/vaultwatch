package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/vaultwatch/internal/alert"
	"github.com/your-org/vaultwatch/internal/alert/sender"
	"github.com/your-org/vaultwatch/internal/config"
	"github.com/your-org/vaultwatch/internal/monitor"
	"github.com/your-org/vaultwatch/internal/vault"
)

func main() {
	cfgPath := flag.String("config", "vaultwatch.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, err := vault.NewClient(cfg.VaultAddress, cfg.VaultToken)
	if err != nil {
		log.Fatalf("failed to create vault client: %v", err)
	}

	senders := buildSenders(cfg)
	notifier := alert.New(senders)

	runner := monitor.NewRunner(client, notifier, cfg.PollInterval, monitor.Thresholds{
		Warning:  cfg.Thresholds.Warning,
		Critical: cfg.Thresholds.Critical,
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go runner.Run()
	log.Println("vaultwatch started")
	<-sig
	log.Println("shutting down")
}

func buildSenders(cfg *config.Config) []sender.Sender {
	var out []sender.Sender
	for _, s := range cfg.Senders {
		switch s.Type {
		case "log":
			out = append(out, sender.NewLogSender())
		case "webhook":
			out = append(out, sender.NewWebhookSender(s.WebhookURL))
		case "slack":
			out = append(out, sender.NewSlackSender(s.WebhookURL))
		case "pagerduty":
			out = append(out, sender.NewPagerDutySender(s.APIKey, s.RoutingKey))
		case "opsgenie":
			out = append(out, sender.NewOpsGenieSender(s.APIKey, s.Team))
		case "victorops":
			out = append(out, sender.NewVictorOpsSender(s.APIKey, s.RoutingKey))
		default:
			log.Printf("unknown sender type %q, skipping", s.Type)
		}
	}
	if len(out) == 0 {
		out = append(out, sender.NewLogSender())
	}
	return out
}
