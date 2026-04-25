package sender

func NewGooglePubSubSenderWithBase(projectID, topicID, apiKey, baseURL string) *googlePubSubSender {
	return newGooglePubSubSenderWithBase(projectID, topicID, apiKey, baseURL)
}
