package main

import (
	"flag"
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
	cfgPath := flag.String("config", "vaultwatch.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, err := vault.NewClient(cfg.Vault.Address, cfg.Vault.Token)
	if err != nil {
		log.Fatalf("failed to create vault client: %v", err)
	}

	senders := buildSenders(cfg)
	notifier := alert.New(senders...)

	leaseMonitor := monitor.New(
		cfg.Monitor.WarnThreshold,
		cfg.Monitor.CriticalThreshold,
	)

	runner := monitor.NewRunner(client, leaseMonitor, notifier, cfg.Monitor.PollInterval)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("vaultwatch started")
	go runner.Run()
	<-sigCh
	log.Println("vaultwatch shutting down")
}

func buildSenders(cfg *config.Config) []alert.Sender {
	var senders []alert.Sender

	for _, sc := range cfg.Alerting.Senders {
		switch sc.Type {
		case "log":
			senders = append(senders, sender.NewLogSender())
		case "webhook":
			senders = append(senders, sender.NewWebhookSender(sc.Params["url"]))
		case "slack":
			senders = append(senders, sender.NewSlackSender(sc.Params["webhook_url"]))
		case "pagerduty":
			senders = append(senders, sender.NewPagerDutySender(sc.Params["routing_key"]))
		case "opsgenie":
			senders = append(senders, sender.NewOpsGenieSender(sc.Params["api_key"]))
		case "datadog":
			senders = append(senders, sender.NewDatadogSender(sc.Params["api_key"], sc.Params["app_key"]))
		case "teams":
			senders = append(senders, sender.NewTeamsSender(sc.Params["webhook_url"]))
		case "discord":
			senders = append(senders, sender.NewDiscordSender(sc.Params["webhook_url"]))
		case "azure_monitor":
			if cfg.Alerting.AzureMonitor != nil {
				am := cfg.Alerting.AzureMonitor
				logType := am.LogType
				if logType == "" {
					logType = "VaultWatch"
				}
				senders = append(senders, sender.NewAzureMonitorSender(am.WorkspaceID, am.SharedKey, logType))
			}
		default:
			log.Printf("unknown sender type %q, skipping", sc.Type)
		}
	}

	if len(senders) == 0 {
		log.Println("no senders configured, defaulting to log sender")
		senders = append(senders, sender.NewLogSender())
	}

	return senders
}
