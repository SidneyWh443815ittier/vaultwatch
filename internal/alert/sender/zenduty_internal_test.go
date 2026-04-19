package sender

func NewZendutySenderWithURL(integrationKey, url string) Sender {
	return newZendutySenderWithURL(integrationKey, url)
}
