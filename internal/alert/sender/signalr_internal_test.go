package sender

import "net/http"

// NewSignalRSenderWithURL exposes the internal constructor for testing.
func NewSignalRSenderWithURL(webhookURL string, client *http.Client) Sender {
	return newSignalRSenderWithURL(webhookURL, client)
}
