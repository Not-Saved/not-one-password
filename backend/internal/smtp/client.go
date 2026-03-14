package smtp

import (
	"main/internal/config"
	"net/smtp"
)

type SMTPClient struct {
	address  string
	from     string
	password string
	host     string
}

func NewSMTPclient(cfg config.SMTPConfig) *SMTPClient {
	address := cfg.Host + ":" + cfg.Port

	return &SMTPClient{
		host:     cfg.Host,
		address:  address,
		from:     cfg.From,
		password: cfg.Password,
	}
}

func (c *SMTPClient) SendEmail(to, subject, body string) error {
	message := c.BuildEmail(to, subject, body)
	var auth smtp.Auth

	if c.password != "" {
		auth = smtp.PlainAuth("", c.from, c.password, c.host)
	}
	return smtp.SendMail(
		c.address,
		auth,
		c.from,
		[]string{to},
		message,
	)
}

func (c *SMTPClient) BuildEmail(to, subject, body string) []byte {
	msg := ""
	msg += "From: " + c.from + "\r\n"
	msg += "To: " + to + "\r\n"
	msg += "Subject: " + subject + "\r\n"
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += body

	return []byte(msg)
}
