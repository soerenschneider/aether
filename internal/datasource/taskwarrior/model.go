package taskwarrior

import (
	"fmt"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

type TaskTemplateData struct {
	// the tasks that are shown, overdue first and limited to the configured amount
	Tasks []Task
	// the shown tasks grouped by due date
	Groups []TaskGroup

	// the counts include tasks that are hidden due to the limit
	Open        int
	Overdue     int
	DueToday    int
	DueTomorrow int
	// amount of tasks that are not shown due to the limit
	Hidden int

	// the task to do first: the longest overdue one, otherwise the next one due today
	Pressing *TaskView
	// the first shown task
	Top *TaskView
}

type Task struct {
	Id          int32
	Description string
	Due         time.Time
	Project     string
	Urgency     float32
	Status      string
	Wait        time.Time
	Tags        []string
}

func parseDate(input string) (time.Time, error) {
	layout := "20060102T150405Z"
	return time.Parse(layout, input)
}

func formatDueTime(dueTime time.Time) string {
	if dueTime.IsZero() {
		return ""
	}

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
		addendum = fmt.Sprintf(" (%v left)", humanized)
	} else {
		humanized := pkg.DurationToString(time.Since(dueTime))
		addendum = fmt.Sprintf(" (%v ago)", humanized)
	}
	if dueTime.Year() == now.Year() {
		return dueTime.Format("02.01.") + addendum
	}

	return dueTime.Format("02.01.2006") + addendum
}
