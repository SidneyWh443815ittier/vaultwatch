package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type signalRSender struct {
	webhookURL string
	client     *http.Client
}

type signalRPayload struct {
	Text string `json:"text"`
}

// NewSignalRSender creates a sender that posts alerts to a SignalR webhook URL.
func NewSignalRSender(webhookURL string) Sender {
	return newSignalRSenderWithURL(webhookURL, &http.Client{})
}

func newSignalRSenderWithURL(webhookURL string, client *http.Client) Sender {
	return &signalRSender{webhookURL: webhookURL, client: client}
}

func (s *signalRSender) Send(level, message string) error {
	payload := signalRPayload{
		Text: fmt.Sprintf("[%s] %s", level, message),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("signalr: marshal payload: %w", err)
	}
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("signalr: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("signalr: unexpected status %d", resp.StatusCode)
	}
	return nil
}
