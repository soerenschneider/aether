package serve

import (
	"context"
	"errors"

	"gopkg.in/gomail.v2"
)

type Email struct {
	server     string
	port       int
	from       string
	recipients []string

	username string
	password string
}

type EmailOpt func(*Email) error

func WithPort(port int) EmailOpt {
	return func(e *Email) error {
		if port < 1 {
			return errors.New("invalid port given")
		}
		e.port = port
		return nil
	}
}

func NewEmail(from string, recipients []string, server, username, password string, opts ...EmailOpt) (*Email, error) {
	e := &Email{
		server:     server,
		from:       from,
		username:   username,
		password:   password,
		port:       465,
		recipients: recipients,
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	return e, nil
}

func (e *Email) SendReport(_ context.Context, subject, html, plain string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", e.from)
	m.SetHeader("To", e.recipients...)
	m.SetHeader("Subject", subject)

	m.SetBody("text/plain", plain)
	m.AddAlternative("text/html", html)

	d := gomail.NewDialer(e.server, e.port, e.username, e.password)
	return d.DialAndSend(m)
}
