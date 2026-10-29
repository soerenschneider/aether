package taskwarrior

import (
	"errors"
	"os"
)

func WithLimit(limit int) Opt {
	return func(datasource *Datasource) error {
		if limit < 1 {
			return errors.New("limit can not be < 1")
		}

		datasource.limit = limit
		return nil
	}
}

func WithSummaryDays(days int) Opt {
	return func(datasource *Datasource) error {
		if days < 1 {
			return errors.New("days can not be < 1")
		}

		datasource.summaryDays = days
		return nil
	}
}

// WithTemplateFile replaces the default template with the template in the given file.
func WithTemplateFile(file string) Opt {
	return func(ds *Datasource) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		ds.customTemplate = data
		return nil
	}
}
