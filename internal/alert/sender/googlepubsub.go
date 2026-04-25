package sender

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultPubSubBase = "https://pubsub.googleapis.com"

type pubSubMessage struct {
	Messages []pubSubEntry `json:"messages"`
}

type pubSubEntry struct {
	Data       string            `json:"data"`
	Attributes map[string]string `json:"attributes"`
}

type pubSubPayload struct {
	Level     string `json:"level"`
	LeaseID   string `json:"lease_id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type googlePubSubSender struct {
	projectID string
	topicID   string
	apiKey    string
	baseURL   string
	client    *http.Client
}

// NewGooglePubSubSender creates a sender that publishes alerts to a Google Cloud Pub/Sub topic.
func NewGooglePubSubSender(projectID, topicID, apiKey string) *googlePubSubSender {
	return newGooglePubSubSenderWithBase(projectID, topicID, apiKey, defaultPubSubBase)
}

func newGooglePubSubSenderWithBase(projectID, topicID, apiKey, baseURL string) *googlePubSubSender {
	return &googlePubSubSender{
		projectID: projectID,
		topicID:   topicID,
		apiKey:    apiKey,
		baseURL:   baseURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *googlePubSubSender) Send(ctx context.Context, level, leaseID, message string) error {
	payload := pubSubPayload{
		Level:     level,
		LeaseID:   leaseID,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("googlepubsub: marshal payload: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(raw)
	body := pubSubMessage{
		Messages: []pubSubEntry{
			{
				Data: encoded,
				Attributes: map[string]string{
					"level":    level,
					"lease_id": leaseID,
				},
			},
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("googlepubsub: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/topics/%s:publish", s.baseURL, s.projectID, s.topicID)
	if s.apiKey != "" {
		url += "?key=" + s.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("googlepubsub: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("googlepubsub: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("googlepubsub: unexpected status %d", resp.StatusCode)
	}
	return nil
}
