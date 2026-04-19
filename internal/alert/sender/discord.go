package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultDiscordURL = "https://discord.com/api/webhooks"

type discordSender struct {
	webhookURL string
	client     *http.Client
}

type discordPayload struct {
	Content  string         `json:"content,omitempty"`
	Embeds   []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
}

// NewDiscordSender creates a sender that posts alerts to a Discord webhook.
func NewDiscordSender(webhookURL string) alert.Sender {
	return newDiscordSenderWithURL(webhookURL, &http.Client{})
}

func newDiscordSenderWithURL(webhookURL string, client *http.Client) alert.Sender {
	return &discordSender{webhookURL: webhookURL, client: client}
}

func (d *discordSender) Send(a alert.Alert) error {
	payload := discordPayload{
		Embeds: []discordEmbed{
			{
				Title:       fmt.Sprintf("[%s] Vault Lease Expiring: %s", a.Level, a.LeaseID),
				Description: fmt.Sprintf("Expires in %s\nPath: %s", a.TTL, a.Path),
				Color:       discordColor(a.Level),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func discordColor(level alert.Level) int {
	switch level {
	case alert.LevelCritical:
		return 0xFF0000 // red
	case alert.LevelWarning:
		return 0xFFA500 // orange
	default:
		return 0x00FF00 // green
	}
}
