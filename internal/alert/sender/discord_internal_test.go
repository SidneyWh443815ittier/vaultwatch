package sender

import "net/http"

// NewDiscordSenderWithURL exposes the internal constructor for testing.
func NewDiscordSenderWithURL(webhookURL string, client *http.Client) interface{ Send(interface{}) error } {
	return newDiscordSenderWithURL(webhookURL, client)
}
