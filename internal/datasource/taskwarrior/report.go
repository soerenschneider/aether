package taskwarrior

import (
	"fmt"
	"time"
)

func GenerateReport(tasks []Task, now time.Time, addSummaryForNoEvents bool) []string {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Track affected tasks
	var overdueTasks []Task
	var dueTodayTasks []Task
	var dueIn7DaysTasks []Task

	// Loop through tasks and categorize them
	for _, task := range tasks {
		if task.Due.IsZero() {
			continue // Ignore uninitialized dates
		}

		// Normalize due date (remove time part)
		dueDate := time.Date(task.Due.Year(), task.Due.Month(), task.Due.Day(), 0, 0, 0, 0, task.Due.Location())

		// Categorize the task
		if dueDate.Before(today) {
			overdueTasks = append(overdueTasks, task)
		} else if dueDate.Equal(today) {
			dueTodayTasks = append(dueTodayTasks, task)
		} else if dueDate.After(today) && dueDate.Before(today.Add(7*24*time.Hour)) {
			dueIn7DaysTasks = append(dueIn7DaysTasks, task)
		}
	}

	var report []string
	if len(overdueTasks) > 0 {
		if len(overdueTasks) == 1 {
			report = append(report, fmt.Sprintf("📋 1 Overdue task: %q", overdueTasks[0].Description))
		} else {
			report = append(report, fmt.Sprintf("📋 %d tasks overdue", len(overdueTasks)))
		}
	}

	if len(dueTodayTasks) > 0 {
		if len(dueTodayTasks) == 1 {
			report = append(report, fmt.Sprintf("📋 1 Task due today: %q", dueTodayTasks[0].Description))
		} else {
			report = append(report, fmt.Sprintf("📋 %d tasks due today", len(dueTodayTasks)))
		}
	}

	if len(dueIn7DaysTasks) > 0 {
		if len(dueIn7DaysTasks) == 1 {
			report = append(report, fmt.Sprintf("📋 1 Task due within the next 7 days: %q", dueIn7DaysTasks[0].Description))
		} else {
			report = append(report, fmt.Sprintf("📋 %d tasks due within the next 7 days", len(dueIn7DaysTasks)))
		}
	}

	if addSummaryForNoEvents && len(report) == 0 {
		report = append(report, "✅ No tasks due next 7d")
	}

	return report
}
