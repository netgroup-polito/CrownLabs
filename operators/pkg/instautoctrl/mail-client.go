package instautoctrl

import (
	"net/smtp"
	"strconv"
)

type MailClient struct {
	SMTPServer string
	SMTPPort   int
	Auth       smtp.Auth
	From       string
}

func (m *MailClient) SendMail(to []string, subject string, body string) error {
	msg := []byte("Subject: " + subject + "\n\n" + body)
	return smtp.SendMail(m.SMTPServer+":"+strconv.Itoa(m.SMTPPort), m.Auth, m.From, to, msg)
}
