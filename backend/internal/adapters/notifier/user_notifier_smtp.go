package notifier

import (
	"main/internal/smtp"
)

type UserNotifierSMTP struct {
	smtp *smtp.SMTPClient
}

func NewUserNotifierSMTP(smtp *smtp.SMTPClient) *UserNotifierSMTP {
	return &UserNotifierSMTP{
		smtp: smtp,
	}
}

func (c *UserNotifierSMTP) NotifyRegistrationIntent(to, code string) error {
	return c.smtp.SendEmail(to, "Registration", code)
}

func (c *UserNotifierSMTP) NotifyRegistrationSuccess(to string) error {
	return c.smtp.SendEmail(to, "Registration Success", "Successfully registered!")
}
