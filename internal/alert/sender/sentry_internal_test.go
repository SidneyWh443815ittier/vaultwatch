package sender

func NewSentrySenderWithURL(authToken, org, project, baseURL string) *sentrySender {
	return newSentrySenderWithURL(authToken, org, project, baseURL)
}
