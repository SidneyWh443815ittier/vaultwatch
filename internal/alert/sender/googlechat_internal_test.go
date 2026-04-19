package sender

import "net/http"

// NewGoogleChatSenderWithURL exposes the internal constructor for testing.
func NewGoogleChatSenderWithURL(webhookURL string, client *http.Client) *googleChatSender {
	return newGoogleChatSenderWithURL(webhookURL, client).(*googleChatSender)
}
