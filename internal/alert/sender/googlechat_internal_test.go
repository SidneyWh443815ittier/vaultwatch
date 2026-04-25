package sender

import "net/http"

// NewGoogleChatSenderWithURL exposes the internal constructor for testing.
// It allows tests to inject a custom HTTP client (e.g. a mock or test server client)
// and a specific webhook URL without going through environment configuration.
func NewGoogleChatSenderWithURL(webhookURL string, client *http.Client) *googleChatSender {
	return newGoogleChatSenderWithURL(webhookURL, client).(*googleChatSender)
}

// WebhookURL returns the webhook URL configured on the sender.
// Exposed for test assertions to verify the URL was set correctly.
func (g *googleChatSender) WebhookURL() string {
	return g.webhookURL
}
