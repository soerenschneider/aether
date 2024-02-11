package taskwarrior

import (
	"bytes"
	"context"
	"html/template"
	"sort"
	"sync"

	"github.com/jubnzv/go-taskwarrior"
	"go.uber.org/multierr"
)

const (
	defaultLimit      = 10
	defaultTaskRcFile = "~/.taskrc"
)

type Opt func(datasource *TaskwarriorDatasource) error

type TaskwarriorDatasource struct {
	limit      int
	template   *template.Template
	once       sync.Once
	taskRcFile string
}

func New(opts ...Opt) (*TaskwarriorDatasource, error) {
	tw := &TaskwarriorDatasource{
		limit:      defaultLimit,
		taskRcFile: defaultTaskRcFile,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(tw); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return tw, errs
}

func (a *TaskwarriorDatasource) Name() string {
	return "Taskwarrior"
}

func (t *TaskwarriorDatasource) GetHtml(_ context.Context) (string, error) {
	t.once.Do(func() {
		if t.template == nil {
			t.template = template.Must(template.New("taskwarrior").Parse(defaultTemplate))
		}
	})

	tasks, err := t.getTaskwarriorOutput()
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.template.Execute(&tpl, tasks)
	return tpl.String(), err
}

func (t *TaskwarriorDatasource) getTaskwarriorOutput() ([]Task, error) {
	tw, _ := taskwarrior.NewTaskWarrior(t.taskRcFile)
	if err := tw.FetchAllTasks(); err != nil {
		return nil, err
	}

	tasks := convertTasks(tw.Tasks)

	filtered := filterTasks(tasks)
	sortTasks(filtered)

	if t.limit > 0 {
		return filtered[:t.limit], nil
	}
	return filtered, nil
}

func filterTask(task Task) bool {
	return task.Status == "pending"
}

func filterTasks(tasks []Task) []Task {
	var filtered []Task
	for _, task := range tasks {
		if filterTask(task) {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

func sortTasks(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Urgency > tasks[j].Urgency
	})
}
