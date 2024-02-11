package weather

import (
	"errors"
	"html/template"
	"net/http"
	"os"
)

func WithTemplateFile(file string) Opt {
	return func(ds *WeatherDatasource) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		temp, err := template.New("weather").Parse(string(data))
		if err != nil {
			return err
		}

		ds.template = temp
		return nil
	}
}

func WithHttpClient(client *http.Client) Opt {
	return func(ds *WeatherDatasource) error {
		if client == nil {
			return errors.New("empty http client provided")
		}

		ds.httpClient = client
		return nil
	}
}
