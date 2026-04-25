package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

type googleChatSender struct {
	webhookURL string
	client     *http.Client
}

type googleChatPayload struct {
	Text string `json:"text"`
}

// NewGoogleChatSender creates a sender that posts messages to a Google Chat webhook.
func NewGoogleChatSender(webhookURL string) alert.Sender {
	return newGoogleChatSenderWithURL(webhookURL)
}

func newGoogleChatSenderWithURL(webhookURL string) *googleChatSender {
	return &googleChatSender{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *googleChatSender) Send(a alert.Alert) error {
	body := googleChatPayload{
		Text: fmt.Sprintf("*[%s] VaultWatch Alert*\nLease: `%s`\nExpires in: %s\n%s",
			string(a.Level),
			a.LeaseID,
			a.TTL.Round(time.Second),
			a.Message,
		),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("googlechat: marshal payload: %w", err)
	}

	resp, err := g.client.Post(g.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("googlechat: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("googlechat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
