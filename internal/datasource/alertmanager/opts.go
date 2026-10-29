package alertmanager

import (
	"errors"
	"html/template"
	"os"
	"strings"

	"github.com/prometheus/alertmanager/api/v2/client/alert"
)

// WithClient replaces the alertmanager API client, e.g. with a fake for development.
func WithClient(client alert.ClientService) Opt {
	return func(ds *AlertmanagerDatasource) error {
		if client == nil {
			return errors.New("nil client passed")
		}
		ds.client = client
		return nil
	}
}

func WithBasePath(path string) Opt {
	return func(ds *AlertmanagerDatasource) error {
		ds.basePath = path
		return nil
	}
}

func WithScheme(scheme string) Opt {
	return func(ds *AlertmanagerDatasource) error {
		scheme = strings.ToLower(scheme)
		if scheme != "http" && scheme != "https" {
			return errors.New("scheme must be either https or http")
		}
		ds.scheme = scheme
		return nil
	}
}

func WithTemplateFile(file string) Opt {
	return func(ds *AlertmanagerDatasource) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		temp, err := template.New("alertmanager").Parse(string(data))
		if err != nil {
			return err
		}

		ds.defaultTemplate = temp
		return nil
	}
}
