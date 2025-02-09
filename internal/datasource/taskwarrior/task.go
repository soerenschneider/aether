package taskwarrior

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"go.uber.org/multierr"
)

const (
	defaultLimit = 10
)

type Client interface {
	GetTasks(ctx context.Context) ([]Task, error)
}

type Opt func(datasource *Datasource) error

type Datasource struct {
	limit           int
	defaultTemplate *template.Template
	simpleTemplate  *template.Template
	taskRcFile      string
	client          Client
}

func New(client Client, templateData templates.TemplateData, opts ...Opt) (*Datasource, error) {
	if client == nil {
		return nil, errors.New("nil client passed")
	}

	if err := templateData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template data: %w", err)
	}

	tw := &Datasource{
		client:     client,
		limit:      defaultLimit,
		taskRcFile: defaultTaskRcFile,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(tw); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	var err error
	tw.defaultTemplate, err = template.New("taskwarrior-default").Funcs(
		map[string]any{
			"formatDue":     formatDueTime,
			"formatProject": formatProject,
			"formatTags":    formatTags,
			"getCssClass":   getCssClass,
		}).Parse(string(templateData.DefaultTemplate))

	if len(templateData.SimpleTemplate) > 0 {
		tw.simpleTemplate, err = template.New("taskwarrior-simple").Funcs(
			map[string]any{
				"formatDue":     formatDueTime,
				"formatProject": formatProject,
				"formatTags":    formatTags,
				"getCssClass":   getCssClass,
			}).Parse(string(templateData.SimpleTemplate))
	}

	return tw, err
}

func (t *Datasource) Name() string {
	return "Taskwarrior"
}

func (t *Datasource) GetData(ctx context.Context) (*internal.Data, error) {
	tasks, err := t.client.GetTasks(ctx)
	if err != nil {
		return nil, err
	}

	tasks = filterTasks(tasks)
	sortTasks(tasks)
	if t.limit > 0 && len(tasks) > t.limit {
		tasks = tasks[0:t.limit]
	}

	data := TaskTemplateData{Tasks: tasks}

	var defaultTemplateRendered bytes.Buffer
	if err := t.defaultTemplate.Execute(&defaultTemplateRendered, data); err != nil {
		return nil, err
	}

	var simpleTemplateRendered bytes.Buffer
	if t.simpleTemplate != nil {
		if err := t.simpleTemplate.Execute(&simpleTemplateRendered, data); err != nil {
			return nil, err
		}
	}

	return &internal.Data{
		Summary:                    GenerateReport(tasks),
		RenderedDefaultTemplate:    defaultTemplateRendered.Bytes(),
		RenderedSimplifiedTemplate: simpleTemplateRendered.Bytes(),
	}, err
}

func filterTask(task Task, now time.Time) bool {
	if !task.Wait.IsZero() && now.Before(task.Wait) {
		return false
	}
	return task.Status == "pending"
}

func filterTasks(tasks []Task) []Task {
	var filtered []Task
	now := time.Now()
	for _, task := range tasks {
		if filterTask(task, now) {
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

func formatTags(tags []string) string {
	return strings.Join(tags, ", ")
}

func formatProject(project string) string {
	return strings.Replace(project, ".", " / ", -1)
}

func getCssClass(dueTime time.Time) string {
	if dueTime.IsZero() {
		return ""
	}

	until := time.Until(dueTime)
	if until <= time.Hour*24 {
		return "red"
	}
	if until <= 3*time.Hour*24 {
		return "orange"
	}
	if until <= 7*time.Hour*24 {
		return "yellow"
	}

	return ""
}
