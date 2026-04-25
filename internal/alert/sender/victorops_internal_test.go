package sender

func NewVictorOpsSenderWithURL(url, routingKey string) Sender {
	return newVictorOpsSenderWithURL(url, routingKey)
}
