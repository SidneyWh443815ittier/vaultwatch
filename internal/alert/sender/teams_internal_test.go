package sender

import "net/http"

// NewTeamsSenderWithURL exposes the internal constructor for testing.
func NewTeamsSenderWithURL(webhookURL string, client *http.Client) interface{ Send(interface{}) error } {
	return newTeamsSenderWithURL(webhookURL, client)
}
