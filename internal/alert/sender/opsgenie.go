package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultOpsGenieURL = "https://api.opsgenie.com/v2/alerts"

type opsGenieSender struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type opsGeniePayload struct {
	Message     string            `json:"message"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Tags        []string          `json:"tags"`
	Details     map[string]string `json:"details"`
}

// NewOpsGenieSender creates an OpsGenie alert sender.
func NewOpsGenieSender(apiKey string) *opsGenieSender {
	return newOpsGenieSenderWithURL(apiKey, defaultOpsGenieURL)
}

func newOpsGenieSenderWithURL(apiKey, url string) *opsGenieSender {
	return &opsGenieSender{
		apiKey:  apiKey,
		baseURL: url,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *opsGenieSender) Send(a alert.Alert) error {
	payload := opsGeniePayload{
		Message:     fmt.Sprintf("[%s] Vault lease expiring: %s", a.Level, a.LeaseID),
		Description: fmt.Sprintf("Lease %s expires in %s", a.LeaseID, a.TTL),
		Priority:    ogPriority(a.Level),
		Tags:        []string{"vaultwatch", string(a.Level)},
		Details: map[string]string{
			"lease_id": a.LeaseID,
			"ttl":      a.TTL.String(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func ogPriority(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "P1"
	case alert.LevelWarning:
		return "P3"
	default:
		return "P5"
	}
}
