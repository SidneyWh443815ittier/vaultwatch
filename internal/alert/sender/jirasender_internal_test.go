package sender

func NewJiraSenderWithURL(url, project, issueType, username, token string) interface {
	Send(interface{ GetLevel() string }) error
} {
	return newJiraSenderWithURL(url, project, issueType, username, token)
}
