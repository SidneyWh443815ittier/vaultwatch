package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultTeamsURL = "https://outlook.office.com/webhook/"

type teamsSender struct {
	webhookURL string
	client     *http.Client
}

type teamsPayload struct {
	Type       string `json:"@type"`
	Context    string `json:"@context"`
	ThemeColor string `json:"themeColor"`
	Summary    string `json:"summary"`
	Text       string `json:"text"`
}

// NewTeamsSender creates a sender that posts alerts to a Microsoft Teams webhook.
func NewTeamsSender(webhookURL string) alert.Sender {
	return newTeamsSenderWithURL(webhookURL, &http.Client{})
}

func newTeamsSenderWithURL(webhookURL string, client *http.Client) alert.Sender {
	return &teamsSender{webhookURL: webhookURL, client: client}
}

func (s *teamsSender) Send(a alert.Alert) error {
	color := teamsColor(a.Level)
	payload := teamsPayload{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: color,
		Summary:    fmt.Sprintf("VaultWatch: %s", a.LeaseID),
		Text:       fmt.Sprintf("**%s** — Lease `%s` expires in %s", a.Level, a.LeaseID, a.TTL),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func teamsColor(level alert.Level) string {
	switch level {
	case alert.LevelCritical:
		return "FF0000"
	case alert.LevelWarning:
		return "FFA500"
	default:
		return "00FF00"
	}
}
