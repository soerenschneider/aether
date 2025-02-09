package alertmanager

import (
	"errors"
	"html/template"
	"os"
	"strings"
)

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
