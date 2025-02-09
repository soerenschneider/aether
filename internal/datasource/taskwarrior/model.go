package taskwarrior

import (
	"fmt"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

type TaskTemplateData struct {
	Tasks []Task
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
