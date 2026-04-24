package sender

func NewStatuspageSenderWithURL(apiKey, pageID, componentID, baseURL string) *statuspageSender {
	return newStatuspageSenderWithURL(apiKey, pageID, componentID, baseURL)
}

func StatuspageStatus(level interface{ String() string }) string {
	return ""
}
