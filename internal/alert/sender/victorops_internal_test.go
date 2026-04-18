package sender

// NewVictorOpsSenderWithURL exposes the internal constructor for testing.
func NewVictorOpsSenderWithURL(url, routingKey string) Sender {
	return newVictorOpsSenderWithURL(url, routingKey)
}
