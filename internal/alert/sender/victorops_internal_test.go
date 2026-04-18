package sender

// NewVictorOpsSenderWithURL exposes the internal constructor for testing.
// It allows tests to override the VictorOps REST endpoint URL, which is
// useful for pointing at a local test server instead of the real API.
func NewVictorOpsSenderWithURL(url, routingKey string) Sender {
	return newVictorOpsSenderWithURL(url, routingKey)
}
