package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

type mattermostSender struct {
	url    string
	client *http.Client
}

type mattermostPayload struct {
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
}

// NewMattermostSender creates a Sender that posts to a Mattermost incoming webhook.
func NewMattermostSender(webhookURL string) Sender {
	return newMattermostSenderWithURL(webhookURL)
}

func newMattermostSenderWithURL(url string) Sender {
	return &mattermostSender{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *mattermostSender) Send(level Level, lease monitor.LeaseInfo) error {
	text := fmt.Sprintf("[%s] Vault lease `%s` expires in %s",
		level, lease.LeaseID, lease.TTL.Round(time.Second))

	payload := mattermostPayload{
		Text:     text,
		Username: "vaultwatch",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mattermost: marshal payload: %w", err)
	}

	resp, err := m.client.Post(m.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mattermost: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mattermost: unexpected status %d", resp.StatusCode)
	}
	return nil
}
