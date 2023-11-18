package taskwarrior

import (
	"bytes"
	"context"
	"html/template"
	"sort"
	"sync"

	"github.com/jubnzv/go-taskwarrior"
)

const defaultLimit = 10

type Opt func(datasource *TaskwarriorDatasource) error

type TaskwarriorDatasource struct {
	limit    int
	template *template.Template
	once     sync.Once
}

func New() (*TaskwarriorDatasource, error) {
	return &TaskwarriorDatasource{
		limit: defaultLimit,
	}, nil
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
	tw, _ := taskwarrior.NewTaskWarrior("~/.taskrc")
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
	if task.Status == "pending" {
		return true
	}

	return false
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
