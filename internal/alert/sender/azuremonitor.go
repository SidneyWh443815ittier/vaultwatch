package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultAzureMonitorURL = "https://management.azure.com"

type azureMonitorSender struct {
	workspaceID string
	sharedKey   string
	logType     string
	url         string
	client      *http.Client
}

type azureMonitorPayload struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	LeaseID   string `json:"lease_id"`
	Timestamp string `json:"timestamp"`
}

// NewAzureMonitorSender creates a sender that posts alerts to Azure Monitor Log Analytics.
func NewAzureMonitorSender(workspaceID, sharedKey, logType string) *azureMonitorSender {
	return newAzureMonitorSenderWithURL(workspaceID, sharedKey, logType, defaultAzureMonitorURL)
}

func newAzureMonitorSenderWithURL(workspaceID, sharedKey, logType, baseURL string) *azureMonitorSender {
	return &azureMonitorSender{
		workspaceID: workspaceID,
		sharedKey:   sharedKey,
		logType:     logType,
		url:         baseURL,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *azureMonitorSender) Send(level, message, leaseID string) error {
	payload := []azureMonitorPayload{
		{
			Level:     level,
			Message:   message,
			LeaseID:   leaseID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("azuremonitor: marshal payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/logs", s.url)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("azuremonitor: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Log-Type", s.logType)
	req.Header.Set("x-ms-date", time.Now().UTC().Format(http.TimeFormat))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("azuremonitor: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("azuremonitor: unexpected status %d", resp.StatusCode)
	}
	return nil
}
