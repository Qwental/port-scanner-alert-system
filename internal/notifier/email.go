package notifier

import (
	"fmt"
	"net/smtp"

	"go.uber.org/zap"
)

type EmailNotifier struct {
	host     string
	port     int
	user     string
	password string
	from     string
	log      *zap.SugaredLogger
}

func NewEmailNotifier(host string, port int, user, password, from string, log *zap.SugaredLogger) *EmailNotifier {
	return &EmailNotifier{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
		log:      log,
	}
}

func (e *EmailNotifier) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", e.host, e.port)

	auth := smtp.PlainAuth("", e.user, e.password, e.host)

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		e.from, to, subject,
	)

	msg := []byte(headers + body)

	err := smtp.SendMail(addr, auth, e.from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("send email failed: %w", err)
	}

	e.log.Infof("Email sent to %s", to)
	return nil
}
