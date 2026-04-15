package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookSender posts alert payloads to an HTTP endpoint.
type WebhookSender struct {
	URL    string
	client *http.Client
}

type webhookPayload struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// NewWebhookSender returns a WebhookSender with a sensible timeout.
func NewWebhookSender(url string) *WebhookSender {
	return &WebhookSender{
		URL:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send marshals the alert and POSTs it to the configured URL.
func (w *WebhookSender) Send(level, message string) error {
	payload := webhookPayload{
		Level:     level,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
