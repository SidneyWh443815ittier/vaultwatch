package sender

import "github.com/yourusername/vaultwatch/internal/alert"

func NewOpsGenieSenderWithURL(apiKey, apiURL string) alert.Sender {
	return newOpsGenieSenderWithURL(apiKey, apiURL)
}
