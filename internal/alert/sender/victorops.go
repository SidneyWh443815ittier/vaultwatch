package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

type victorOpsSender struct {
	postURL string
	client  *http.Client
}

type victorOpsPayload struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
	MonitoringTool    string `json:"monitoring_tool"`
}

// NewVictorOpsSender creates a VictorOps (Splunk On-Call) sender.
func NewVictorOpsSender(routingKey, integrationURL string) *victorOpsSender {
	url := fmt.Sprintf("%s/%s", integrationURL, routingKey)
	return newVictorOpsSenderWithURL(url)
}

func newVictorOpsSenderWithURL(url string) *victorOpsSender {
	return &victorOpsSender{
		postURL: url,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *victorOpsSender) Send(a alert.Alert) error {
	payload := victorOpsPayload{
		MessageType:       voMessageType(a.Level),
		EntityID:          a.LeaseID,
		EntityDisplayName: fmt.Sprintf("Vault lease expiring: %s", a.LeaseID),
		StateMessage:      fmt.Sprintf("Lease %s has TTL %s remaining", a.LeaseID, a.TTL),
		MonitoringTool:    "vaultwatch",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("victorops: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.postURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("victorops: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("victorops: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("victorops: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func voMessageType(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "CRITICAL"
	case alert.LevelWarning:
		return "WARNING"
	default:
		return "INFO"
	}
}
