package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultStatuspageURL = "https://api.statuspage.io/v1"

type statuspageSender struct {
	apiKey     string
	pageID     string
	componentID string
	baseURL    string
	client     *http.Client
}

type statuspageComponentBody struct {
	Component statuspageComponentStatus `json:"component"`
}

type statuspageComponentStatus struct {
	Status string `json:"status"`
}

func statuspageStatus(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "major_outage"
	case alert.LevelWarning:
		return "degraded_performance"
	default:
		return "operational"
	}
}

func NewStatuspageSender(apiKey, pageID, componentID string) *statuspageSender {
	return newStatuspageSenderWithURL(apiKey, pageID, componentID, defaultStatuspageURL)
}

func newStatuspageSenderWithURL(apiKey, pageID, componentID, baseURL string) *statuspageSender {
	return &statuspageSender{
		apiKey:      apiKey,
		pageID:      pageID,
		componentID: componentID,
		baseURL:     baseURL,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *statuspageSender) Send(msg alert.Message) error {
	body := statuspageComponentBody{
		Component: statuspageComponentStatus{
			Status: statuspageStatus(msg.Level),
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("statuspage: marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/pages/%s/components/%s", s.baseURL, s.pageID, s.componentID)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("statuspage: create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "OAuth "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("statuspage: request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("statuspage: unexpected status %d", resp.StatusCode)
	}
	return nil
}
