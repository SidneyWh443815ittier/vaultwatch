package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultGoogleChatURL = "https://chat.googleapis.com/v1/spaces"

type googleChatSender struct {
	webhookURL string
	client     *http.Client
}

type googleChatPayload struct {
	Text string `json:"text"`
}

// NewGoogleChatSender creates a sender that posts alerts to a Google Chat webhook.
func NewGoogleChatSender(webhookURL string) Sender {
	return newGoogleChatSenderWithURL(webhookURL)
}

func newGoogleChatSenderWithURL(webhookURL string) *googleChatSender {
	return &googleChatSender{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *googleChatSender) Send(alert Alert) error {
	body := googleChatPayload{
		Text: fmt.Sprintf("[%s] %s — %s (expires in %s)",
			alert.Level, alert.LeaseID, alert.Message, alert.TTL),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("googlechat: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("googlechat: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("googlechat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
