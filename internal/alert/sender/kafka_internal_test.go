package sender

import "net/http"

// NewKafkaSenderWithURL exposes the internal constructor for white-box tests.
func NewKafkaSenderWithURL(url, topic string, client *http.Client) Sender {
	return newKafkaSenderWithURL(url, topic, client)
}
