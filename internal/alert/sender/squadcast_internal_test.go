package sender

func NewSquadcastSenderWithURL(webhookURL string) *squadcastSender {
	return &squadcastSender{
		webhookURL: webhookURL,
		client:     newTestHTTPClient(),
	}
}
