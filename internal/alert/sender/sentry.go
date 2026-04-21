package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultSentryURL = "https://sentry.io/api/0/projects"

type sentrySender struct {
	dsn        string
	project    string
	org        string
	authToken  string
	baseURL    string
	httpClient *http.Client
}

type sentryPayload struct {
	Title     string            `json:"title"`
	Level     string            `json:"level"`
	Platform  string            `json:"platform"`
	Timestamp string            `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
	Extra     map[string]string `json:"extra"`
}

func sentryLevel(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "fatal"
	case alert.LevelWarning:
		return "warning"
	default:
		return "info"
	}
}

func NewSentrySender(authToken, org, project string) *sentrySender {
	return newSentrySenderWithURL(authToken, org, project, defaultSentryURL)
}

func newSentrySenderWithURL(authToken, org, project, baseURL string) *sentrySender {
	return &sentrySender{
		authToken:  authToken,
		org:        org,
		project:    project,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *sentrySender) Send(msg alert.Message) error {
	payload := sentryPayload{
		Title:     msg.LeaseID,
		Level:     sentryLevel(msg.Level),
		Platform:  "other",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Tags:      map[string]string{"lease_id": msg.LeaseID},
		Extra:     map[string]string{"ttl": msg.TTL.String(), "body": msg.Body},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("sentry: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s/store/", s.baseURL, s.org, s.project)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("sentry: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.authToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sentry: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sentry: unexpected status %d", resp.StatusCode)
	}
	return nil
}
