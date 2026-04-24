package sender

func NewOpsGenieSenderWithURL(apiKey, url string) *opsGenieSender {
	return newOpsGenieSenderWithURL(apiKey, url)
}
