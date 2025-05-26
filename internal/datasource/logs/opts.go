package logs

import (
	"errors"
	"html/template"
	"net/http"
	"os"
)

func WithTemplateFile(file string) Opt {
	return func(ds *VictorialogsClient) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		temp, err := template.New("anniversaries").Parse(string(data))
		if err != nil {
			return err
		}

		ds.regularTemplate = temp
		return nil
	}
}

func WithLimit(limit int) Opt {
	return func(ds *VictorialogsClient) error {
		if limit < 1 || limit > 50 {
			return errors.New("limit must be within range [1, 50]")
		}

		ds.limit = limit
		return nil
	}
}

func WithQuery(query string) Opt {
	return func(ds *VictorialogsClient) error {
		if query == "" {
			return errors.New("empty query supplied")
		}

		ds.query = query
		return nil
	}
}

func WithHttpClient(client *http.Client) Opt {
	return func(ds *VictorialogsClient) error {
		if client == nil {
			return errors.New("empty http client provided")
		}

		ds.httpClient = client
		return nil
	}
}
