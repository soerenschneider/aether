package stocks

import (
	"errors"
	"html/template"
	"net/http"
	"os"
)

func WithClient(client *http.Client) Opts {
	return func(datasource *StocksDatasource) error {
		if client == nil {
			return errors.New("empty httpClient provided")
		}

		datasource.httpClient = client
		return nil
	}
}

func WithTemplateFile(file string) Opts {
	return func(ds *StocksDatasource) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		temp, err := template.New("stocks").Parse(string(data))
		if err != nil {
			return err
		}

		ds.template = temp
		return nil
	}
}

func WithHttpClient(client *http.Client) Opts {
	return func(ds *StocksDatasource) error {
		if client == nil {
			return errors.New("empty http httpClient provided")
		}

		ds.httpClient = client
		return nil
	}
}
