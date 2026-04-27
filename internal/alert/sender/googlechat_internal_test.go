package sender

import "net/http"

// NewGoogleChatSenderWithURL exposes the internal constructor for testing.
func NewGoogleChatSenderWithURL(webhookURL string) *googleChatSender {
	return newGoogleChatSenderWithURL(webhookURL)
}

// SetClient replaces the HTTP client on a googleChatSender (test helper).
func (s *googleChatSender) SetClient(c *http.Client) {
	s.client = c
}
