package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultDatadogURL = "https://api.datadoghq.com/api/v1/events"

type datadogSender struct {
	apiKey  string
	url     string
	client  *http.Client
}

type datadogPayload struct {
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	AlertType string   `json:"alert_type"`
	Tags      []string `json:"tags,omitempty"`
}

// NewDatadogSender creates a Sender that posts events to the Datadog Events API.
func NewDatadogSender(apiKey string) alert.Sender {
	return newDatadogSenderWithURL(apiKey, defaultDatadogURL)
}

func newDatadogSenderWithURL(apiKey, url string) alert.Sender {
	return &datadogSender{
		apiKey: apiKey,
		url:    url,
		client: &http.Client{},
	}
}

func (d *datadogSender) Send(a alert.Alert) error {
	payload := datadogPayload{
		Title:     fmt.Sprintf("VaultWatch: %s", a.LeaseID),
		Text:      fmt.Sprintf("Lease %s expires in %s (status: %s)", a.LeaseID, a.TTL, a.Level),
		AlertType: ddAlertType(a.Level),
		Tags:      []string{"source:vaultwatch", fmt.Sprintf("lease:%s", a.LeaseID)},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("datadog: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, d.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("datadog: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("datadog: post event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("datadog: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func ddAlertType(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "error"
	case alert.LevelWarning:
		return "warning"
	default:
		return "info"
	}
}
