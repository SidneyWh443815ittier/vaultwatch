package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultAmplitudeURL = "https://api2.amplitude.com/2/httpapi"

type amplitudeSender struct {
	apiKey string
	url    string
	client *http.Client
}

type amplitudePayload struct {
	APIKey string           `json:"api_key"`
	Events []amplitudeEvent `json:"events"`
}

type amplitudeEvent struct {
	UserID      string                 `json:"user_id"`
	EventType   string                 `json:"event_type"`
	EventProperties map[string]interface{} `json:"event_properties"`
}

// NewAmplitudeSender creates a sender that logs vault lease alert events to Amplitude.
func NewAmplitudeSender(apiKey string) (alert.Sender, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("amplitude: api_key is required")
	}
	return newAmplitudeSenderWithURL(apiKey, defaultAmplitudeURL), nil
}

func newAmplitudeSenderWithURL(apiKey, url string) alert.Sender {
	return &amplitudeSender{
		apiKey: apiKey,
		url:    url,
		client: &http.Client{},
	}
}

func (a *amplitudeSender) Send(msg alert.Message) error {
	payload := amplitudePayload{
		APIKey: a.apiKey,
		Events: []amplitudeEvent{
			{
				UserID:    "vaultwatch",
				EventType: "vault_lease_alert",
				EventProperties: map[string]interface{}{
					"level":     msg.Level,
					"lease_id":  msg.LeaseID,
					"expires_in": msg.ExpiresIn.String(),
					"summary":   msg.Summary,
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("amplitude: marshal payload: %w", err)
	}

	resp, err := a.client.Post(a.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("amplitude: post event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("amplitude: unexpected status %d", resp.StatusCode)
	}
	return nil
}
