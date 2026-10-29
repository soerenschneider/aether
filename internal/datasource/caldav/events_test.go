package caldav

import (
	"strings"
	"testing"
	"time"
	// makes the tests independent of the time zone database of the system
	_ "time/tzdata"

	"github.com/emersion/go-ical"
)

func decodeCalendar(t *testing.T, lines ...string) *ical.Calendar {
	t.Helper()
	data := strings.Join(append(append([]string{"BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//aether//test//EN"}, lines...), "END:VCALENDAR", ""), "\r\n")
	cal, err := ical.NewDecoder(strings.NewReader(data)).Decode()
	if err != nil {
		t.Fatal(err)
	}
	return cal
}

func berlin(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

type wantEntry struct {
	summary    string
	start, end string
}

func assertEntries(t *testing.T, got []Entry, want []wantEntry) {
	t.Helper()
	sortEntries(got)
	if len(got) != len(want) {
		var summaries []string
		for _, e := range got {
			summaries = append(summaries, e.Summary+" "+e.Start.Format("02.01. 15:04"))
		}
		t.Fatalf("expected %d entries, got %d: %v", len(want), len(got), summaries)
	}
	for i, w := range want {
		start, end := got[i].Start.Format("2006-01-02 15:04"), got[i].End.Format("2006-01-02 15:04")
		if got[i].Summary != w.summary || start != w.start || end != w.end {
			t.Errorf("entry %d: got {%s %s %s}, want %+v", i, got[i].Summary, start, end, w)
		}
	}
}

func TestEntriesFromCalendar_recurring(t *testing.T) {
	loc := berlin(t)
	cal := decodeCalendar(t,
		"BEGIN:VEVENT",
		"UID:swim@test",
		"DTSTART;TZID=Europe/Berlin:20261013T170000",
		"DTEND;TZID=Europe/Berlin:20261013T180000",
		"RRULE:FREQ=WEEKLY;COUNT=8",
		// a list of dates in a single property
		"EXDATE;TZID=Europe/Berlin:20261103T170000,20261110T170000",
		"RDATE;TZID=Europe/Berlin:20261031T100000",
		`SUMMARY:Swimming\, kids`,
		"LOCATION:Stadtbad",
		"END:VEVENT",
		// moved occurrence
		"BEGIN:VEVENT",
		"UID:swim@test",
		"RECURRENCE-ID;TZID=Europe/Berlin:20261020T170000",
		"DTSTART;TZID=Europe/Berlin:20261021T180000",
		"DTEND;TZID=Europe/Berlin:20261021T190000",
		`SUMMARY:Swimming\, kids (moved)`,
		"END:VEVENT",
		// cancelled occurrence
		"BEGIN:VEVENT",
		"UID:swim@test",
		"RECURRENCE-ID;TZID=Europe/Berlin:20261117T170000",
		"DTSTART;TZID=Europe/Berlin:20261117T170000",
		"DTEND;TZID=Europe/Berlin:20261117T180000",
		"STATUS:CANCELLED",
		`SUMMARY:Swimming\, kids`,
		"END:VEVENT",
	)

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	to := time.Date(2026, 12, 31, 0, 0, 0, 0, loc)
	entries := EntriesFromCalendar(cal, from, to, loc)

	assertEntries(t, entries, []wantEntry{
		{"Swimming, kids", "2026-10-13 17:00", "2026-10-13 18:00"},
		{"Swimming, kids (moved)", "2026-10-21 18:00", "2026-10-21 19:00"},
		// daylight saving time ended on the 25th, the event keeps its local time
		{"Swimming, kids", "2026-10-27 17:00", "2026-10-27 18:00"},
		{"Swimming, kids", "2026-10-31 10:00", "2026-10-31 11:00"},
		{"Swimming, kids", "2026-11-24 17:00", "2026-11-24 18:00"},
		{"Swimming, kids", "2026-12-01 17:00", "2026-12-01 18:00"},
	})
	if entries[0].Location != "Stadtbad" || len(entries[0].Formatted) == 0 {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestEntriesFromCalendar_allDayYearly(t *testing.T) {
	loc := berlin(t)
	cal := decodeCalendar(t,
		"BEGIN:VEVENT",
		"UID:birthday@test",
		"DTSTART;VALUE=DATE:19900130",
		"DTEND;VALUE=DATE:19900131",
		"RRULE:FREQ=YEARLY",
		"SUMMARY:Maria's birthday",
		"END:VEVENT",
	)

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	to := time.Date(2027, 10, 1, 0, 0, 0, 0, loc)
	entries := EntriesFromCalendar(cal, from, to, loc)

	assertEntries(t, entries, []wantEntry{
		{"Maria's birthday", "2027-01-30 00:00", "2027-01-31 00:00"},
	})
	if !entries[0].isAllDay() {
		t.Error("expected an all-day entry")
	}
}

func TestEntriesFromCalendar_singleEvents(t *testing.T) {
	loc := berlin(t)
	cal := decodeCalendar(t,
		// in UTC, as some servers store events
		"BEGIN:VEVENT",
		"UID:utc@test",
		"DTSTART:20261029T090000Z",
		"DTEND:20261029T093000Z",
		"SUMMARY:Standup",
		"END:VEVENT",
		// floating time, i.e. local time
		"BEGIN:VEVENT",
		"UID:floating@test",
		"DTSTART:20261029T120000",
		"DURATION:PT1H30M",
		"SUMMARY:Lunch",
		"END:VEVENT",
		"BEGIN:VEVENT",
		"UID:cancelled@test",
		"DTSTART:20261029T150000Z",
		"STATUS:CANCELLED",
		"SUMMARY:Cancelled meeting",
		"END:VEVENT",
		// an invalid rule still shows the event once
		"BEGIN:VEVENT",
		"UID:broken@test",
		"DTSTART:20261030T080000Z",
		"DTEND:20261030T090000Z",
		"RRULE:FREQ=SOMETIMES",
		"SUMMARY:Broken rule",
		"END:VEVENT",
	)

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	to := time.Date(2026, 12, 31, 0, 0, 0, 0, loc)
	assertEntries(t, EntriesFromCalendar(cal, from, to, loc), []wantEntry{
		{"Standup", "2026-10-29 10:00", "2026-10-29 10:30"},
		{"Lunch", "2026-10-29 12:00", "2026-10-29 13:30"},
		{"Broken rule", "2026-10-30 09:00", "2026-10-30 10:00"},
	})
}

func TestEntriesFromCalendar_window(t *testing.T) {
	loc := berlin(t)
	cal := decodeCalendar(t,
		"BEGIN:VEVENT",
		"UID:daily@test",
		"DTSTART;TZID=Europe/Berlin:20000101T070000",
		"DTEND;TZID=Europe/Berlin:20000101T073000",
		"RRULE:FREQ=DAILY",
		"SUMMARY:Morning run",
		"END:VEVENT",
	)

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	to := time.Date(2026, 11, 1, 0, 0, 0, 0, loc)
	entries := EntriesFromCalendar(cal, from, to, loc)
	if len(entries) != 31 {
		t.Fatalf("expected an occurrence on each day of october, got %d", len(entries))
	}
	sortEntries(entries)
	if first := entries[0].Start.Format("2006-01-02 15:04"); first != "2026-10-01 07:00" {
		t.Errorf("unexpected first occurrence %s", first)
	}
	if last := entries[30].Start.Format("2006-01-02 15:04"); last != "2026-10-31 07:00" {
		t.Errorf("unexpected last occurrence %s", last)
	}
}
