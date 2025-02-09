package taskwarrior

import (
	"cmp"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/go-taskwarrior"
)

const defaultTaskRcFile = "~/.taskrc"

type TaskwarriorClient struct {
	taskRcFile string
}

func NewTaskwarriorClient(taskRcFile string) (*TaskwarriorClient, error) {
	return &TaskwarriorClient{
		taskRcFile: cmp.Or(taskRcFile, defaultTaskRcFile),
	}, nil
}

func (t *TaskwarriorClient) GetTasks(_ context.Context) ([]Task, error) {
	tw, _ := taskwarrior.NewTaskWarrior(t.taskRcFile)
	if err := tw.FetchAllTasks(); err != nil {
		return nil, err
	}

	return convertTasks(tw.Tasks), nil
}

func convertTasks(tasks []taskwarrior.Task) []Task {
	ret := make([]Task, len(tasks))
	for index, task := range tasks {
		ret[index] = convertTask(task)
	}
	return ret
}

func convertTask(t taskwarrior.Task) Task {
	ret := Task{
		Id:          t.Id,
		Description: t.Description,
		Project:     t.Project,
		Urgency:     t.Urgency,
		Status:      t.Status,
		Tags:        t.Tags,
	}

	if len(t.Due) > 0 {
		dueTime, err := parseDate(t.Due)
		if err != nil {
			log.Warn().Err(err).Msg("could not parse time from task")
			return ret
		}
		ret.Due = dueTime
	}

	if len(t.Wait) > 0 {
		waitTime, err := parseDate(t.Wait)
		if err != nil {
			log.Warn().Err(err).Msg("could not parse time from task")
			return ret
		}
		ret.Wait = waitTime
	}

	return ret
}
