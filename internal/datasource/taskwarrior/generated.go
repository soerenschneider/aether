package taskwarrior

import (
	"fmt"
	"strings"
	"time"

	"github.com/soerenschneider/aether/pkg"

	"github.com/jubnzv/go-taskwarrior"
	"github.com/rs/zerolog/log"
)

type Task struct {
	Id          int32
	Description string
	Due         string
	DueCssClass string
	Project     string
	Urgency     float32
	Status      string
}

func convertTask(t taskwarrior.Task) Task {
	ret := Task{
		Id:          t.Id,
		Description: t.Description,
		Due:         t.Due,
		Project:     translateProject(t.Project),
		Urgency:     t.Urgency,
		Status:      t.Status,
	}

	if len(t.Due) > 0 {
		dueTime, err := parseDueTime(t.Due)
		if err != nil {
			log.Warn().Err(err).Msg("could not parse time from task")
			return ret
		}
		ret.Due = formatDueTime(dueTime)
		ret.DueCssClass = getCssClass(dueTime)
	}

	return ret
}

func translateProject(project string) string {
	return strings.Replace(project, ".", " / ", -1)
}

func convertTasks(tasks []taskwarrior.Task) []Task {
	ret := make([]Task, len(tasks))
	for index, task := range tasks {
		ret[index] = convertTask(task)
	}
	return ret
}

func getCssClass(dueTime time.Time) string {
	if dueTime.IsZero() {
		return ""
	}

	until := time.Until(dueTime)
	if until > 7*time.Hour*24 {
		return "yellow"
	}
	if until > 3*time.Hour*24 {
		return "orange"
	}
	if until <= 0 {
		return "red"
	}
	return "green"
}

func parseDueTime(input string) (time.Time, error) {
	layout := "20060102T150405Z"
	return time.Parse(layout, input)
}

func formatDueTime(dueTime time.Time) string {
	now := time.Now()

	if pkg.IsToday(dueTime, now) {
		return "Today"
	}

	if pkg.IsTomorrow(dueTime) {
		return "Tomorrow"
	}

	var addendum string

	if dueTime.After(now) {
		humanized := pkg.DurationToString(time.Until(dueTime))
		addendum = fmt.Sprintf(" (in %v)", humanized)
	} else {
		humanized := pkg.DurationToString(time.Since(dueTime))
		addendum = fmt.Sprintf(" (%v ago)", humanized)
	}
	if dueTime.Year() == now.Year() {
		return dueTime.Format("02.01.") + addendum
	}

	return dueTime.Format("02.01.") + addendum
}
