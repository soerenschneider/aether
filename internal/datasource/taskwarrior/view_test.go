package taskwarrior

import (
	"testing"
	"time"
)

var (
	cet = time.FixedZone("CET", 3600)
	// Thursday, a fixed point in time so the tests don't depend on when they run
	testNow = time.Date(2026, 10, 29, 10, 0, 0, 0, cet)
)

func at(days, hour, minute int) time.Time {
	return time.Date(2026, 10, 29+days, hour, minute, 0, 0, cet)
}

func Test_dueBucketAndLabel(t *testing.T) {
	tests := []struct {
		name       string
		due        time.Time
		wantBucket string
		wantLabel  string
	}{
		{"no due date", time.Time{}, bucketNone, ""},
		{"three days ago", at(-3, 9, 0), bucketOverdue, "3 days overdue"},
		{"yesterday", at(-1, 0, 0), bucketOverdue, "1 day overdue"},
		{"today, earlier", at(0, 8, 0), bucketToday, "today 08:00"},
		{"today, no time", at(0, 0, 0), bucketToday, "today"},
		{"today, later", at(0, 18, 0), bucketToday, "today 18:00"},
		// taskwarrior reports UTC: midnight in CET is 23:00 UTC on the day before
		{"today, reported in UTC", at(0, 0, 0).UTC(), bucketToday, "today"},
		{"tomorrow", at(1, 9, 0), bucketTomorrow, "tomorrow 09:00"},
		{"this week", at(3, 0, 0), bucketLater, "Sun"},
		{"in two weeks", at(14, 0, 0), bucketLater, "Thu 12.11."},
		{"next year", time.Date(2027, 2, 1, 0, 0, 0, 0, cet), bucketLater, "01.02.2027"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dueBucket(tt.due, testNow); got != tt.wantBucket {
				t.Errorf("dueBucket() = %q, want %q", got, tt.wantBucket)
			}
			if got := dueLabel(tt.due, testNow); got != tt.wantLabel {
				t.Errorf("dueLabel() = %q, want %q", got, tt.wantLabel)
			}
		})
	}
}

func Test_newTemplateData(t *testing.T) {
	tasks := []Task{
		{Id: 1, Description: "urgent, no due date", Urgency: 20},
		{Id: 8, Description: "less urgent, no due date", Urgency: 10},
		{Id: 2, Description: "later", Due: at(5, 0, 0), Urgency: 15},
		{Id: 3, Description: "overdue, low urgency", Due: at(-1, 0, 0), Urgency: 1},
		{Id: 4, Description: "overdue since long", Due: at(-10, 0, 0), Urgency: 2},
		{Id: 5, Description: "today, evening", Due: at(0, 18, 0), Urgency: 9},
		{Id: 6, Description: "today, noon", Due: at(0, 12, 0), Urgency: 3},
		{Id: 7, Description: "tomorrow", Due: at(1, 0, 0), Urgency: 5},
	}

	data := newTemplateData(tasks, 5, testNow)

	if data.Open != 8 || data.Overdue != 2 || data.DueToday != 2 || data.DueTomorrow != 1 {
		t.Errorf("unexpected counts: open=%d overdue=%d today=%d tomorrow=%d", data.Open, data.Overdue, data.DueToday, data.DueTomorrow)
	}
	if data.Hidden != 3 {
		t.Errorf("expected 3 hidden tasks, got %d", data.Hidden)
	}

	// tasks without a due date are sorted by urgency
	all := newTemplateData(tasks, 0, testNow)
	last := all.Groups[len(all.Groups)-1]
	if last.Label != "No due date" || last.Tasks[0].Id != 1 || last.Tasks[1].Id != 8 {
		t.Errorf("unexpected order of tasks without due date: %+v", last)
	}

	// overdue first despite their low urgency, by due date within each group; the limit cuts the least pressing
	wantIds := []int32{4, 3, 6, 5, 7}
	if len(data.Tasks) != len(wantIds) {
		t.Fatalf("expected %d tasks, got %d", len(wantIds), len(data.Tasks))
	}
	for i, id := range wantIds {
		if data.Tasks[i].Id != id {
			t.Errorf("task %d: got id %d, want %d", i, data.Tasks[i].Id, id)
		}
	}

	wantGroups := []string{"Overdue", "Today", "Tomorrow"}
	if len(data.Groups) != len(wantGroups) {
		t.Fatalf("expected %d groups, got %d", len(wantGroups), len(data.Groups))
	}
	for i, label := range wantGroups {
		if data.Groups[i].Label != label {
			t.Errorf("group %d: got %q, want %q", i, data.Groups[i].Label, label)
		}
	}

	if data.Pressing == nil || data.Pressing.Id != 4 {
		t.Errorf("expected the longest overdue task to be the most pressing, got %v", data.Pressing)
	}
	if data.Top == nil || data.Top.Id != 4 {
		t.Errorf("expected task 4 on top, got %v", data.Top)
	}
}

func Test_newTemplateData_pressingWithoutOverdue(t *testing.T) {
	tasks := []Task{
		{Id: 1, Description: "today, evening", Due: at(0, 18, 0), Urgency: 9},
		{Id: 2, Description: "today, noon", Due: at(0, 12, 0), Urgency: 3},
		{Id: 3, Description: "tomorrow", Due: at(1, 0, 0), Urgency: 5},
	}

	data := newTemplateData(tasks, 0, testNow)
	if data.Pressing == nil || data.Pressing.Id != 2 {
		t.Errorf("expected the next task due today to be the most pressing, got %v", data.Pressing)
	}

	data = newTemplateData(tasks[2:], 0, testNow)
	if data.Pressing != nil {
		t.Errorf("expected no pressing task, got %v", data.Pressing)
	}
	if data.DueTomorrow != 1 || data.Top == nil || data.Top.Id != 3 {
		t.Errorf("unexpected data: %+v", data)
	}
}
