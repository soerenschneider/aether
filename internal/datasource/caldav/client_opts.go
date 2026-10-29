package caldav

import (
	"errors"
	"net/http"
)

func WithBasicAuth(username, password string) ClientOpt {
	return func(ds *Client) error {
		ds.username = username
		ds.password = password
		return nil
	}
}

// WithCalendar selects a calendar of the account by its name or path, instead of the first one.
func WithCalendar(calendar string) ClientOpt {
	return func(ds *Client) error {
		ds.calendar = calendar
		return nil
	}
}

func WithHttpClient(client *http.Client) ClientOpt {
	return func(ds *Client) error {
		if client == nil {
			return errors.New("empty http client provided")
		}

		ds.httpClient = client
		return nil
	}
}
