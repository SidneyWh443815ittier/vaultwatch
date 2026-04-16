// This file exposes internal helpers for testing.
package sender

import "net/http"
import "time"

// NewPagerDutySenderWithURL creates a PagerDutySender that posts to a custom URL.
// Used in tests to point at a mock HTTP server instead of PagerDuty.
func NewPagerDutySenderWithURL(integrationKey, url string) *PagerDutySender {
	return &PagerDutySender{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: 5 * time.Second},
		eventsURL:      url,
	}
}
