package logs

import (
	"errors"
	"net/http"
	"os"
)

// WithQuerier replaces the component that queries victorialogs, e.g. with a fake for development.
func WithQuerier(querier Querier) Opt {
	return func(ds *VictorialogsClient) error {
		if querier == nil {
			return errors.New("nil querier passed")
		}
		ds.querier = querier
		return nil
	}
}

// WithTemplateFile replaces the default template with the template in the given file.
func WithTemplateFile(file string) Opt {
	return func(ds *VictorialogsClient) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		ds.customTemplate = data
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
