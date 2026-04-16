package sender

import "net/http"

func NewSlackSenderWithURL(webhookURL string, client *http.Client) *SlackSender {
	return newSlackSenderWithURL(webhookURL, client)
}
