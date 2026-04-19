package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultRocketChatURL = "https://your.rocketchat.server/hooks"

type rocketChatSender struct {
	webhookURL string
	client     *http.Client
}

type rocketChatPayload struct {
	Text        string `json:"text"`
	Attachments []rocketChatAttachment `json:"attachments,omitempty"`
}

type rocketChatAttachment struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Color string `json:"color"`
}

// NewRocketChatSender creates a Sender that posts to a Rocket.Chat incoming webhook.
func NewRocketChatSender(webhookURL string) alert.Sender {
	return newRocketChatSenderWithURL(webhookURL, &http.Client{})
}

func newRocketChatSenderWithURL(webhookURL string, client *http.Client) alert.Sender {
	return &rocketChatSender{webhookURL: webhookURL, client: client}
}

func (s *rocketChatSender) Send(a alert.Alert) error {
	color := "#36a64f"
	if a.Level == alert.Warning {
		color = "#ffae42"
	} else if a.Level == alert.Critical {
		color = "#d00000"
	}

	payload := rocketChatPayload{
		Text: fmt.Sprintf("*VaultWatch Alert*: %s", a.LeaseID),
		Attachments: []rocketChatAttachment{
			{
				Title: fmt.Sprintf("[%s] Lease expiring", a.Level),
				Text:  fmt.Sprintf("Lease `%s` expires in %s", a.LeaseID, a.TTL),
				Color: color,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rocketchat: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("rocketchat: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rocketchat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
