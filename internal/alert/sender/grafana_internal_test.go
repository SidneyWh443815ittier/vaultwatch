package sender

func NewGrafanaSenderWithURL(url, apiKey string) *grafanaSender {
	return newGrafanaSenderWithURL(url, apiKey)
}
