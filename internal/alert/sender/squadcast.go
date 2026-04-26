package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultSquadcastURL = "https://api.squadcast.com/v2/incidents/api"

type squadcastSender struct {
	webhookURL string
	client     *http.Client
}

type squadcastPayload struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Tags        map[string]string `json:"tags,omitempty"`
	Status      string `json:"status"`
}

// NewSquadcastSender returns a Sender that posts incidents to Squadcast.
func NewSquadcastSender(webhookURL string) alert.Sender {
	return newSquadcastSenderWithURL(webhookURL, defaultSquadcastURL)
}

func newSquadcastSenderWithURL(webhookURL, apiURL string) alert.Sender {
	_ = apiURL // webhookURL is the full endpoint in Squadcast
	return &squadcastSender{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *squadcastSender) Send(a alert.Alert) error {
	payload := squadcastPayload{
		Message:     fmt.Sprintf("[%s] Vault lease expiring: %s", a.Level, a.LeaseID),
		Description: fmt.Sprintf("Lease %s expires in %s (TTL: %s)", a.LeaseID, a.TimeLeft, a.TTL),
		Status:      squadcastStatus(a.Level),
		Tags: map[string]string{
			"level":    string(a.Level),
			"lease_id": a.LeaseID,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("squadcast: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("squadcast: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("squadcast: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func squadcastStatus(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "trigger"
	case alert.LevelWarning:
		return "trigger"
	default:
		return "resolve"
	}
}
