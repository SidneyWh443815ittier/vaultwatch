package sender

import "net/http"

// NewGoogleChatSenderWithURL exposes the internal constructor for tests.
func NewGoogleChatSenderWithURL(webhookURL string) *googleChatSender {
	return newGoogleChatSenderWithURL(webhookURL)
}

// SetHTTPClient allows tests to inject a custom HTTP client.
func (s *googleChatSender) SetHTTPClient(c *http.Client) {
	s.client = c
}
