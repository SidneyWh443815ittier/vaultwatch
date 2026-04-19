package sender

import "net/http"

// NewRocketChatSenderWithURL exposes the internal constructor for testing.
func NewRocketChatSenderWithURL(webhookURL string, client *http.Client) interface {
	Send(a interface{}) error
} {
	return newRocketChatSenderWithURL(webhookURL, client)
}
