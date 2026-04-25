package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultPrometheusURL = "http://localhost:9091"

type prometheusSender struct {
	baseURL    string
	job        string
	httpClient *http.Client
}

type prometheusPayload struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      *time.Time        `json:"endsAt,omitempty"`
}

// NewPrometheusSender creates a sender that pushes alerts to the Prometheus
// Alertmanager API endpoint.
func NewPrometheusSender(baseURL, job string) Sender {
	return newPrometheusSenderWithURL(baseURL, job)
}

func newPrometheusSenderWithURL(baseURL, job string) *prometheusSender {
	if baseURL == "" {
		baseURL = defaultPrometheusURL
	}
	if job == "" {
		job = "vaultwatch"
	}
	return &prometheusSender{
		baseURL:    baseURL,
		job:        job,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *prometheusSender) Send(level, message string) error {
	payload := []prometheusPayload{
		{
			Labels: map[string]string{
				"alertname": "VaultSecretExpiry",
				"severity":  level,
				"job":       s.job,
			},
			Annotations: map[string]string{
				"summary": message,
			},
			StartsAt: time.Now().UTC(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("prometheus sender: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/alerts", s.baseURL)
	resp, err := s.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("prometheus sender: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("prometheus sender: unexpected status %d", resp.StatusCode)
	}
	return nil
}
