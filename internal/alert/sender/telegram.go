package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/vaultwatch/internal/alert"
)

const defaultTelegramAPIBase = "https://api.telegram.org"

type telegramSender struct {
	token   string
	chatID  string
	apiBase string
	client  *http.Client
}

type telegramPayload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// NewTelegramSender creates a Sender that posts alerts to a Telegram chat.
func NewTelegramSender(token, chatID string) alert.Sender {
	return newTelegramSenderWithBase(token, chatID, defaultTelegramAPIBase)
}

func newTelegramSenderWithBase(token, chatID, apiBase string) alert.Sender {
	return &telegramSender{
		token:   token,
		chatID:  chatID,
		apiBase: apiBase,
		client:  &http.Client{},
	}
}

func (t *telegramSender) Send(a alert.Alert) error {
	text := fmt.Sprintf("*[%s] VaultWatch Alert*\nLease: `%s`\nExpires in: %s\nStatus: %s",
		a.Level, a.LeaseID, a.TTL, a.Message)

	payload := telegramPayload{
		ChatID:    t.chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", t.apiBase, t.token)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}
