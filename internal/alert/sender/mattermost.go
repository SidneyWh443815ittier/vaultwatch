package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

type mattermostSender struct {
	webhookURL string
	client     *http.Client
}

type mattermostPayload struct {
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
}

// NewMattermostSender creates a Sender that posts messages to a Mattermost incoming webhook.
func NewMattermostSender(webhookURL string) alert.Sender {
	return newMattermostSenderWithURL(webhookURL, &http.Client{})
}

func newMattermostSenderWithURL(webhookURL string, client *http.Client) alert.Sender {
	return &mattermostSender{webhookURL: webhookURL, client: client}
}

func (m *mattermostSender) Send(a alert.Alert) error {
	text := fmt.Sprintf("[%s] Lease `%s` expires in %s — %s",
		a.Level, a.LeaseID, a.TTL.Round(1000000000), a.Message)

	payload := mattermostPayload{
		Text:     text,
		Username: "vaultwatch",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mattermost: marshal payload: %w", err)
	}

	resp, err := m.client.Post(m.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mattermost: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mattermost: unexpected status %d", resp.StatusCode)
	}
	return nil
}
