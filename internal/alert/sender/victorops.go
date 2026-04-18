package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultVictorOpsURL = "https://alert.victorops.com/integrations/generic/20131114/alert"

type victorOpsSender struct {
	webhookURL string
	routingKey string
	client     *http.Client
}

type voPayload struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
}

// NewVictorOpsSender creates a Sender that posts alerts to VictorOps.
func NewVictorOpsSender(apiKey, routingKey string) Sender {
	url := fmt.Sprintf("%s/%s/%s", defaultVictorOpsURL, apiKey, routingKey)
	return newVictorOpsSenderWithURL(url, routingKey)
}

func newVictorOpsSenderWithURL(url, routingKey string) Sender {
	return &victorOpsSender{
		webhookURL: url,
		routingKey: routingKey,
		client:     &http.Client{},
	}
}

func (v *victorOpsSender) Send(level, message string) error {
	payload := voPayload{
		MessageType:       voMessageType(level),
		EntityID:          "vaultwatch/lease",
		EntityDisplayName: fmt.Sprintf("VaultWatch [%s]", level),
		StateMessage:      message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("victorops: marshal payload: %w", err)
	}
	resp, err := v.client.Post(v.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("victorops: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("victorops: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func voMessageType(level string) string {
	switch level {
	case "critical":
		return "CRITICAL"
	case "warning":
		return "WARNING"
	default:
		return "INFO"
	}
}
