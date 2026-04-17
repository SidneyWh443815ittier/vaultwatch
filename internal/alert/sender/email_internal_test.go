package sender

import "net/smtp"

// sendMailFunc allows tests to replace smtp.SendMail.
var sendMailFunc = smtp.SendMail

// newEmailSenderWithFunc creates an EmailSender using an injectable send function (test helper).
func newEmailSenderWithFunc(
	host string, port int, user, password, from string, to []string,
	fn func(addr string, a smtp.Auth, from string, to []string, msg []byte) error,
) *EmailSender {
	s := NewEmailSender(host, port, user, password, from, to)
	s.sendFn = fn
	return s
}
