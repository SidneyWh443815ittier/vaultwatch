package sender

import "net/http"

// NewKafkaSenderWithURL exposes the internal constructor for white-box tests.
// It allows tests to inject a custom HTTP client and target URL, bypassing the
// default broker discovery logic used in production.
func NewKafkaSenderWithURL(url, topic string, client *http.Client) Sender {
	return newKafkaSenderWithURL(url, topic, client)
}

// NewKafkaSenderWithDefaults exposes the internal constructor for white-box
// tests using a default http.Client, useful when only the URL and topic need
// to be controlled.
func NewKafkaSenderWithDefaults(url, topic string) Sender {
	return newKafkaSenderWithURL(url, topic, http.DefaultClient)
}
