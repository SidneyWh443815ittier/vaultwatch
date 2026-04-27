package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

type gotifySender struct {
	baseURL  string
	token    string
	client   *http.Client
}

type gotifyPayload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// NewGotifySender creates a Sender that posts notifications to a Gotify server.
func NewGotifySender(baseURL, token string) alert.Sender {
	return newGotifySenderWithURL(baseURL, token)
}

func newGotifySenderWithURL(baseURL, token string) alert.Sender {
	return &gotifySender{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *gotifySender) Send(a alert.Alert) error {
	priority := gotifyPriority(a.Level)

	payload := gotifyPayload{
		Title:    fmt.Sprintf("[%s] VaultWatch: %s", a.Level, a.LeaseID),
		Message:  fmt.Sprintf("Secret %s expires in %s", a.LeaseID, a.TTL),
		Priority: priority,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gotify: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/message?token=%s", g.baseURL, g.token)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gotify: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gotify: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gotify: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func gotifyPriority(level alert.Level) int {
	switch level {
	case alert.LevelCritical:
		return 10
	case alert.LevelWarning:
		return 5
	default:
		return 1
	}
}
