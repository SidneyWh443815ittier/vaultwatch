package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutySender sends alerts to PagerDuty via the Events API v2.
type PagerDutySender struct {
	integrationKey string
	client         *http.Client
}

type pdPayload struct {
	RoutingKey  string     `json:"routing_key"`
	EventAction string     `json:"event_action"`
	Payload     pdDetails  `json:"payload"`
}

type pdDetails struct {
	Summary   string `json:"summary"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

// NewPagerDutySender creates a PagerDutySender with the given integration key.
func NewPagerDutySender(integrationKey string) *PagerDutySender {
	return &PagerDutySender{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts an alert event to PagerDuty.
func (p *PagerDutySender) Send(level, message string) error {
	severity := pdSeverity(level)
	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:   message,
			Source:    "vaultwatch",
			Severity:  severity,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal error: %w", err)
	}
	resp, err := p.client.Post(pagerDutyEventsURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: request error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func pdSeverity(level string) string {
	switch level {
	case "critical":
		return "critical"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}
