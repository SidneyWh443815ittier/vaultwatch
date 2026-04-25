package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

const defaultVictorOpsURL = "https://alert.victorops.com/integrations/generic/20131114/alert"

type victorOpsSender struct {
	url        string
	routingKey string
	client     *http.Client
}

type voPayload struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
	Timestamp         int64  `json:"timestamp"`
}

func voMessageType(level Level) string {
	switch level {
	case LevelCritical:
		return "CRITICAL"
	case LevelWarning:
		return "WARNING"
	default:
		return "INFO"
	}
}

// NewVictorOpsSender creates a Sender that posts alerts to VictorOps.
func NewVictorOpsSender(apiKey, routingKey string) Sender {
	return newVictorOpsSenderWithURL(
		fmt.Sprintf("%s/%s/%s", defaultVictorOpsURL, apiKey, routingKey),
		routingKey,
	)
}

func newVictorOpsSenderWithURL(url, routingKey string) Sender {
	return &victorOpsSender{
		url:        url,
		routingKey: routingKey,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (v *victorOpsSender) Send(level Level, lease monitor.LeaseInfo) error {
	payload := voPayload{
		MessageType:       voMessageType(level),
		EntityID:          lease.LeaseID,
		EntityDisplayName: fmt.Sprintf("Vault lease expiring: %s", lease.LeaseID),
		StateMessage:      fmt.Sprintf("Lease %s expires in %s", lease.LeaseID, lease.TTL),
		Timestamp:         time.Now().Unix(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("victorops: marshal payload: %w", err)
	}

	resp, err := v.client.Post(v.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("victorops: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("victorops: unexpected status %d", resp.StatusCode)
	}
	return nil
}
