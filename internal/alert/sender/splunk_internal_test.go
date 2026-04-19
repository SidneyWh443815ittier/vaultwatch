package sender

import "github.com/yourusername/vaultwatch/internal/alert"

func NewSplunkSenderWithURL(url, token string) alert.Sender {
	return newSplunkSenderWithURL(url, token)
}
