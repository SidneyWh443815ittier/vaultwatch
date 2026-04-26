package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

const defaultCircleCIBaseURL = "https://circleci.com/api/v2"

type circleciSender struct {
	baseURL string
	token   string
	projectSlug string
	httpClient  *http.Client
}

type circleciEnvVarPayload struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type circleciAlertPayload struct {
	Level     string `json:"level"`
	LeaseID   string `json:"lease_id"`
	ExpiresIn string `json:"expires_in"`
	Timestamp string `json:"timestamp"`
}

// NewCircleCISender creates a sender that posts alert metadata to a CircleCI
// pipeline trigger endpoint using a personal API token.
func NewCircleCISender(token, projectSlug string) *circleciSender {
	return newCircleCISenderWithURL(defaultCircleCIBaseURL, token, projectSlug)
}

func newCircleCISenderWithURL(baseURL, token, projectSlug string) *circleciSender {
	return &circleciSender{
		baseURL:     baseURL,
		token:       token,
		projectSlug: projectSlug,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *circleciSender) Send(level string, lease monitor.LeaseInfo) error {
	payload := map[string]interface{}{
		"parameters": map[string]interface{}{
			"vault_alert_level":      level,
			"vault_lease_id":         lease.LeaseID,
			"vault_expires_in_secs":  int(lease.TTL.Seconds()),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("circleci sender: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/project/%s/pipeline", s.baseURL, s.projectSlug)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("circleci sender: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Circle-Token", s.token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("circleci sender: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("circleci sender: unexpected status %d", resp.StatusCode)
	}
	return nil
}
