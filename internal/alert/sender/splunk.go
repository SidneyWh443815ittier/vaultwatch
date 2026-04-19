package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultSplunkURL = "https://localhost:8088/services/collector/event"

type splunkSender struct {
	url   string
	token string
	client *http.Client
}

type splunkEvent struct {
	Time       float64        `json:"time"`
	Sourcetype string         `json:"sourcetype"`
	Event      splunkPayload  `json:"event"`
}

type splunkPayload struct {
	Level   string `json:"level"`
	Lease   string `json:"lease_id"`
	Message string `json:"message"`
}

func NewSplunkSender(hecURL, token string) alert.Sender {
	return newSplunkSenderWithURL(hecURL, token)
}

func newSplunkSenderWithURL(url, token string) alert.Sender {
	if url == "" {
		url = defaultSplunkURL
	}
	return &splunkSender{url: url, token: token, client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *splunkSender) Send(a alert.Alert) error {
	payload := splunkEvent{
		Time:       float64(a.At.Unix()),
		Sourcetype: "vaultwatch",
		Event: splunkPayload{
			Level:   a.Level.String(),
			Lease:   a.LeaseID,
			Message: fmt.Sprintf("Vault lease %s expires in %s", a.LeaseID, a.TTL),
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Splunk "+s.token)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("splunk: unexpected status %d", resp.StatusCode)
	}
	return nil
}
