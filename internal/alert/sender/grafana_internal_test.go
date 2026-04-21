package sender

import "time"

// NewGrafanaSenderWithURL creates a grafanaSender with a custom URL for testing purposes.
func NewGrafanaSenderWithURL(url, apiKey string) *grafanaSender {
	return newGrafanaSenderWithURL(url, apiKey)
}

// NewGrafanaSenderWithURLAndTimeout creates a grafanaSender with a custom URL and
// HTTP timeout for testing purposes.
func NewGrafanaSenderWithURLAndTimeout(url, apiKey string, timeout time.Duration) *grafanaSender {
	s := newGrafanaSenderWithURL(url, apiKey)
	s.httpClient.Timeout = timeout
	return s
}
