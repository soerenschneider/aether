package carddav

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
)

func WithTemplateFile(file string) Opt {
	return func(ds *CarddavDatasource) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		temp, err := template.New("anniversaries").Parse(string(data))
		if err != nil {
			return err
		}

		ds.template = temp
		return nil
	}
}

func WithBasicAuth(username, password string) Opt {
	return func(ds *CarddavDatasource) error {
		ds.username = username
		ds.password = password
		return nil
	}
}

func WithHttpClient(client *http.Client) Opt {
	return func(ds *CarddavDatasource) error {
		if client == nil {
			return errors.New("empty http client provided")
		}

		ds.httpClient = client
		return nil
	}
}

func WithLookaheadDays(days int) Opt {
	return func(ds *CarddavDatasource) error {
		if days < 3 || days > 180 {
			return fmt.Errorf("lookahead days should be [3, 180]")
		}

		ds.lookaheadDays = days
		return nil
	}
}
