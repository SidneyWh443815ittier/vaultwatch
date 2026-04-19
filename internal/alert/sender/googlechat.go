package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

type googleChatSender struct {
	webhookURL string
	client     *http.Client
}

type gchatPayload struct {
	Text string `json:"text"`
}

// NewGoogleChatSender creates a sender that posts messages to a Google Chat webhook.
func NewGoogleChatSender(webhookURL string) alert.Sender {
	return newGoogleChatSenderWithURL(webhookURL, &http.Client{})
}

func newGoogleChatSenderWithURL(webhookURL string, client *http.Client) alert.Sender {
	return &googleChatSender{webhookURL: webhookURL, client: client}
}

func (g *googleChatSender) Send(a alert.Alert) error {
	text := fmt.Sprintf("[%s] Lease *%s* expires in %s — %s",
		a.Level, a.LeaseID, a.TTL, a.Message)

	payload := gchatPayload{Text: text}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("googlechat: marshal payload: %w", err)
	}

	resp, err := g.client.Post(g.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("googlechat: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("googlechat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
