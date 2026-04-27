package sender

import "net/http"

// NewGoogleChatSenderWithURL exposes the internal constructor for testing.
func NewGoogleChatSenderWithURL(webhookURL string, client *http.Client) *googleChatSender {
	s := newGoogleChatSenderWithURL(webhookURL)
	if client != nil {
		s.client = client
	}
	return s
}

// GoogleChatPayload exposes the payload struct for assertion in tests.
type GoogleChatPayload = googleChatPayload
