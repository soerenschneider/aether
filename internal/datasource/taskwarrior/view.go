package taskwarrior

import (
	"fmt"
	"sort"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

// helpers that prepare the tasks for the templates, grouped by calendar day of their due date

const (
	bucketOverdue  = "overdue"
	bucketToday    = "today"
	bucketTomorrow = "tomorrow"
	bucketLater    = "later"
	bucketNone     = "none"
)

var bucketOrder = map[string]int{
	bucketOverdue:  0,
	bucketToday:    1,
	bucketTomorrow: 2,
	bucketLater:    3,
	bucketNone:     4,
}

var bucketLabels = map[string]string{
	bucketOverdue:  "Overdue",
	bucketToday:    "Today",
	bucketTomorrow: "Tomorrow",
	bucketLater:    "Later",
	bucketNone:     "No due date",
}

type TaskView struct {
	Task
	// one of overdue, today, tomorrow, later or none
	Bucket string
	// relative due date, e.g. "2 days overdue", "today 18:00" or "Mon 02.11."
	DueLabel string
}

type TaskGroup struct {
	Label  string
	Bucket string
	Tasks  []TaskView
}

func newTemplateData(tasks []Task, limit int, now time.Time) TaskTemplateData {
	views := make([]TaskView, 0, len(tasks))
	for _, task := range tasks {
		views = append(views, TaskView{
			Task:     task,
			Bucket:   dueBucket(task.Due, now),
			DueLabel: dueLabel(task.Due, now),
		})
	}

	// the most pressing tasks first, so the limit never hides an overdue task in favor of a more urgent later one.
	// within a group by due date, as mixed up dates are hard to read, and by urgency for tasks without a due date
	sort.SliceStable(views, func(i, j int) bool {
		if bucketOrder[views[i].Bucket] != bucketOrder[views[j].Bucket] {
			return bucketOrder[views[i].Bucket] < bucketOrder[views[j].Bucket]
		}
		if !views[i].Due.Equal(views[j].Due) {
			return views[i].Due.Before(views[j].Due)
		}
		return views[i].Urgency > views[j].Urgency
	})

	data := TaskTemplateData{Open: len(views)}
	for i, view := range views {
		switch view.Bucket {
		case bucketOverdue:
			data.Overdue++
			// the task that is overdue for the longest time
			if data.Pressing == nil || data.Pressing.Bucket != bucketOverdue || view.Due.Before(data.Pressing.Due) {
				data.Pressing = &views[i]
			}
		case bucketToday:
			data.DueToday++
			// the next task due today, unless there is an overdue one
			if data.Pressing == nil || (data.Pressing.Bucket == bucketToday && view.Due.Before(data.Pressing.Due)) {
				data.Pressing = &views[i]
			}
		case bucketTomorrow:
			data.DueTomorrow++
		}
	}

	shown := views
	if limit > 0 && len(shown) > limit {
		shown = shown[:limit]
	}
	data.Hidden = len(views) - len(shown)

	for _, view := range shown {
		data.Tasks = append(data.Tasks, view.Task)
		if len(data.Groups) == 0 || data.Groups[len(data.Groups)-1].Bucket != view.Bucket {
			data.Groups = append(data.Groups, TaskGroup{Label: bucketLabels[view.Bucket], Bucket: view.Bucket})
		}
		group := &data.Groups[len(data.Groups)-1]
		group.Tasks = append(group.Tasks, view)
	}

	if len(data.Tasks) > 0 {
		top := data.Groups[0].Tasks[0]
		data.Top = &top
	}

	return data
}

func dueBucket(due, now time.Time) string {
	if due.IsZero() {
		return bucketNone
	}

	days := daysBetween(now, due)
	switch {
	case days < 0:
		return bucketOverdue
	case days == 0:
		return bucketToday
	case days == 1:
		return bucketTomorrow
	default:
		return bucketLater
	}
}

func dueLabel(due, now time.Time) string {
	if due.IsZero() {
		return ""
	}

	due = due.In(now.Location())
	// tasks due "on a day" are due at midnight, there's no point in showing that time
	clock := ""
	if due.Hour() != 0 || due.Minute() != 0 {
		clock = " " + due.Format("15:04")
	}

	days := daysBetween(now, due)
	switch {
	case days < 0:
		if days == -1 {
			return "1 day overdue"
		}
		return fmt.Sprintf("%d days overdue", -days)
	case days == 0:
		return "today" + clock
	case days == 1:
		return "tomorrow" + clock
	case days < 7:
		return due.Format("Mon") + clock
	case due.Year() == now.Year():
		return due.Format("Mon 02.01.")
	default:
		return due.Format("02.01.2006")
	}
}

// daysBetween returns the number of calendar days from now until t, negative if t lies in the past.
func daysBetween(now, t time.Time) int {
	from := pkg.Today(now)
	to := pkg.Today(t.In(now.Location()))
	// round, as days are not always 24 hours long due to daylight saving time
	return int(to.Sub(from).Round(24*time.Hour).Hours() / 24)
}
