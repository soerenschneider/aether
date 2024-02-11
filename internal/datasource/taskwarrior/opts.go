package taskwarrior

import (
	"errors"
	"html/template"
	"os"
)

func WithLimit(limit int) Opt {
	return func(datasource *TaskwarriorDatasource) error {
		if limit < 1 {
			return errors.New("limit can not be < 1")
		}

		datasource.limit = limit
		return nil
	}
}

func WithTaskRcFile(file string) Opt {
	return func(datasource *TaskwarriorDatasource) error {
		if len(file) == 0 {
			return errors.New("no valid file given")
		}

		datasource.taskRcFile = file
		return nil
	}
}

func WithTemplateFile(file string) Opt {
	return func(ds *TaskwarriorDatasource) error {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		temp, err := template.New("taskwarrior").Parse(string(data))
		if err != nil {
			return err
		}

		ds.template = temp
		return nil
	}
}
