package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultNewRelicURL = "https://api.newrelic.com/v2/alerts_events.json"

type newRelicSender struct {
	apiKey string
	url    string
	client *http.Client
}

type nrPayload struct {
	Event nrEvent `json:"event"`
}

type nrEvent struct {
	EventType string `json:"eventType"`
	Summary   string `json:"summary"`
	Level     string `json:"level"`
	LeaseID   string `json:"leaseId"`
	TTL       int64  `json:"ttlSeconds"`
}

func NewNewRelicSender(apiKey string) alert.Sender {
	return newNewRelicSenderWithURL(apiKey, defaultNewRelicURL)
}

func newNewRelicSenderWithURL(apiKey, url string) alert.Sender {
	return &newRelicSender{apiKey: apiKey, url: url, client: &http.Client{}}
}

func (s *newRelicSender) Send(n alert.Notification) error {
	p := nrPayload{
		Event: nrEvent{
			EventType: "VaultLeaseAlert",
			Summary:   n.Message,
			Level:     string(n.Level),
			LeaseID:   n.LeaseID,
			TTL:       int64(n.TTL.Seconds()),
		},
	}
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", s.apiKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("newrelic: unexpected status %d", resp.StatusCode)
	}
	return nil
}
