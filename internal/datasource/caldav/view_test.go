package caldav

import (
	"testing"
	"time"
)

// a fixed point in time, so the tests don't depend on when they run
var testNow = time.Date(2026, 10, 29, 10, 0, 0, 0, time.Local)

func at(days, hour, minute int) time.Time {
	return time.Date(testNow.Year(), testNow.Month(), testNow.Day()+days, hour, minute, 0, 0, time.Local)
}

func entry(summary string, start, end time.Time) Entry {
	return Entry{Summary: summary, Start: start, End: end}
}

func TestCaldavDatasource_filter(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	ds := &CaldavDatasource{location: now.Location(), maxDays: defaultMaxDays, maxEntries: defaultMaxEntries}

	entries := []Entry{
		entry("all day yesterday", today.AddDate(0, 0, -1), today),
		entry("all day today", today, today.AddDate(0, 0, 1)),
		entry("multi-day, started yesterday", today.AddDate(0, 0, -1), today.AddDate(0, 0, 2)),
		entry("later today", today.Add(23*time.Hour), today.Add(23*time.Hour+30*time.Minute)),
		entry("too far away", today.AddDate(0, 0, defaultMaxDays+2), today.AddDate(0, 0, defaultMaxDays+3)),
	}

	got := ds.filter(entries)
	want := []string{"all day today", "multi-day, started yesterday", "later today"}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Summary != want[i] {
			t.Errorf("entry %d: got %q, want %q", i, got[i].Summary, want[i])
		}
	}
}

func TestCaldavData_Days(t *testing.T) {
	data := CaldavData{
		Now: testNow,
		Entries: []Entry{
			entry("trip", at(-1, 0, 0), at(2, 0, 0)),
			entry("offsite", at(0, 0, 0), at(1, 0, 0)),
			entry("run", at(0, 7, 0), at(0, 8, 0)),
			entry("review", at(0, 9, 30), at(0, 10, 30)),
			entry("lunch", at(0, 12, 0), at(0, 13, 0)),
			entry("dinner", at(1, 19, 0), at(1, 21, 30)),
			entry("climbing", at(3, 18, 0), at(3, 20, 0)),
		},
	}

	days := data.Days()
	if len(days) != 3 {
		t.Fatalf("expected 3 days, got %d", len(days))
	}

	wantLabels := []string{"Today", "Tomorrow", at(3, 0, 0).Format("Mon 02.01.")}
	for i, label := range wantLabels {
		if days[i].Label != label {
			t.Errorf("day %d: got label %q, want %q", i, days[i].Label, label)
		}
	}

	type want struct {
		summary, when string
		past, ongoing bool
	}
	wantToday := []want{
		{"trip", "until tomorrow", false, false},
		{"offsite", "All day", false, false},
		{"run", "07:00 – 08:00", true, false},
		{"review", "09:30 – 10:30", false, true},
		{"lunch", "12:00 – 13:00", false, false},
	}
	if len(days[0].Entries) != len(wantToday) {
		t.Fatalf("expected %d entries today, got %d", len(wantToday), len(days[0].Entries))
	}
	for i, w := range wantToday {
		got := days[0].Entries[i]
		if got.Summary != w.summary || got.When != w.when || got.Past != w.past || got.Ongoing != w.ongoing {
			t.Errorf("entry %d: got {%s %q past=%v ongoing=%v}, want %+v", i, got.Summary, got.When, got.Past, got.Ongoing, w)
		}
	}
}

func TestCaldavData_tileHelpers(t *testing.T) {
	data := CaldavData{
		Now: testNow,
		Entries: []Entry{
			entry("offsite", at(0, 0, 0), at(1, 0, 0)),
			entry("review", at(0, 9, 30), at(0, 10, 30)),
			entry("lunch", at(0, 12, 0), at(0, 13, 0)),
			entry("dentist", at(0, 14, 0), at(0, 15, 0)),
			entry("yoga", at(0, 18, 0), at(0, 19, 0)),
			entry("dinner", at(1, 19, 0), at(1, 21, 30)),
		},
	}

	if current := data.Current(); current == nil || current.Summary != "review" {
		t.Errorf("expected review to be current, got %v", current)
	}
	next := data.Next()
	if next == nil || next.Summary != "lunch" {
		t.Fatalf("expected lunch to be next, got %v", next)
	}
	if later := data.LaterToday(); later != 3 {
		t.Errorf("expected 3 entries later today, got %d", later)
	}
	if more := data.MoreToday(); more != 2 {
		t.Errorf("expected 2 more entries today, got %d", more)
	}
	if when := data.ShortWhen(next); when != "12:00" {
		t.Errorf("got %q", when)
	}
	if when := data.ShortWhen(&data.Entries[5]); when != "tomorrow 19:00" {
		t.Errorf("got %q", when)
	}

	// an all-day event is not "current"
	data.Entries = data.Entries[:1]
	if current := data.Current(); current != nil {
		t.Errorf("expected no current entry, got %v", current)
	}
}
