package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackSender struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

func NewSlackSender(webhookURL string) *SlackSender {
	return &SlackSender{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

func newSlackSenderWithURL(webhookURL string, client *http.Client) *SlackSender {
	return &SlackSender{webhookURL: webhookURL, client: client}
}

func (s *SlackSender) Send(level, message string) error {
	text := fmt.Sprintf("[%s] %s", level, message)
	payload := slackPayload{Text: text}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
