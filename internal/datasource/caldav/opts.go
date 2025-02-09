package caldav

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

func WithRegularTemplateFile(file string) Opt {
	return func(ds *CaldavDatasource) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		temp, err := template.New("agenda").Parse(string(data))
		if err != nil {
			return err
		}

		ds.defaultTemplate = temp
		return nil
	}
}

func WithLocation(location *time.Location) Opt {
	return func(ds *CaldavDatasource) error {
		if location == nil {
			return errors.New("empty location")
		}
		ds.location = location
		return nil
	}
}

func WithBasicAuth(username, password string) Opt {
	return func(ds *CaldavDatasource) error {
		ds.username = username
		ds.password = password
		return nil
	}
}

func WithHttpClient(client *http.Client) Opt {
	return func(ds *CaldavDatasource) error {
		if client == nil {
			return errors.New("empty http client provided")
		}

		ds.httpClient = client
		return nil
	}
}

func WithLookaheadDays(days int) Opt {
	return func(ds *CaldavDatasource) error {
		if days < 3 || days > 180 {
			return fmt.Errorf("lookahead days should be [3, 180]")
		}

		ds.maxDays = days
		return nil
	}
}
