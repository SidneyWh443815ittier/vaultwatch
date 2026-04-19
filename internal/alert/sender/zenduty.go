package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultZendutyURL = "https://events.zenduty.com/api/events/"

type zendutyPayload struct {
	AlertType string `json:"alert_type"`
	Message   string `json:"message"`
	Summary   string `json:"summary"`
	EntityID  string `json:"entity_id,omitempty"`
}

type zendutyS struct {
	integrationKey string
	baseURL        string
	client         *http.Client
}

// NewZendutySender creates a Sender that posts alerts to Zenduty.
func NewZendutySender(integrationKey string) Sender {
	return newZendutySenderWithURL(integrationKey, defaultZendutyURL)
}

func newZendutySenderWithURL(integrationKey, baseURL string) Sender {
	return &zendutyS{integrationKey: integrationKey, baseURL: baseURL, client: &http.Client{}}
}

func (z *zendutyS) Send(level, message string) error {
	alertType := zdAlertType(level)
	p := zendutyPayload{
		AlertType: alertType,
		Message:   message,
		Summary:   fmt.Sprintf("VaultWatch: %s", message),
	}
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s%s/", z.baseURL, z.integrationKey)
	resp, err := z.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("zenduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func zdAlertType(level string) string {
	switch level {
	case "critical":
		return "critical"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}
