package sender

func NewGoogleChatSenderWithURL(webhookURL string) *googleChatSender {
	return newGoogleChatSenderWithURL(webhookURL)
}
