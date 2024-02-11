package serve

import (
	"net/smtp"

	"go.uber.org/multierr"
)

type Email struct {
	server     string
	from       string
	recipients []string

	username string
	password string
}

type EmailOpt func(*Email) error

func WithPassword(username, password string) EmailOpt {
	return func(e *Email) error {
		e.username = username
		e.password = password
		return nil
	}
}

func NewEmail(from string, recipients []string, server string, opts ...EmailOpt) (*Email, error) {
	e := &Email{
		server:     server,
		from:       from,
		recipients: recipients,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(e); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return e, errs
}

func (e *Email) getAuth() smtp.Auth {
	return smtp.CRAMMD5Auth(e.username, e.password)
	//return smtp.PlainAuth("", e.from, e.password, e.server)
}

func (e *Email) Send(body string) error {
	subject := "Subject: Test email from Go!\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(subject + mime + body)
	auth := e.getAuth()
	return smtp.SendMail(e.server, auth, e.from, e.recipients, msg)
}
