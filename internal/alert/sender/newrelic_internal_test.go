package sender

import "github.com/yourusername/vaultwatch/internal/alert"

func NewNewRelicSenderWithURL(apiKey, url string) alert.Sender {
	return newNewRelicSenderWithURL(apiKey, url)
}
