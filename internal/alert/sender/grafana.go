package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultGrafanaURL = "https://grafana.example.com/api/alerts"

type grafanaSender struct {
	url    string
	apiKey string
	client *http.Client
}

type grafanaPayload struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	State   string `json:"state"`
}

func grafanaState(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "alerting"
	case alert.LevelWarning:
		return "pending"
	default:
		return "ok"
	}
}

func NewGrafanaSender(apiKey string) *grafanaSender {
	return newGrafanaSenderWithURL(defaultGrafanaURL, apiKey)
}

func newGrafanaSenderWithURL(url, apiKey string) *grafanaSender {
	return &grafanaSender{url: url, apiKey: apiKey, client: &http.Client{}}
}

func (g *grafanaSender) Send(n alert.Notification) error {
	p := grafanaPayload{
		Title:   fmt.Sprintf("VaultWatch: %s", n.LeaseID),
		Message: fmt.Sprintf("Lease %s expires in %s", n.LeaseID, n.TTL),
		State:   grafanaState(n.Level),
	}
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, g.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("grafana: unexpected status %d", resp.StatusCode)
	}
	return nil
}
