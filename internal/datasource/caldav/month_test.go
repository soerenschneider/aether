package caldav

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_buildMonth(t *testing.T) {
	entries := []Entry{
		entry("offsite", at(0, 0, 0), at(1, 0, 0)),
		entry("run", at(0, 7, 0), at(0, 8, 0)),
		entry("review", at(0, 9, 30), at(0, 10, 30)),
		entry("lunch", at(0, 12, 0), at(0, 13, 0)),
		entry("yoga", at(0, 18, 0), at(0, 19, 0)),
		// from friday into november
		entry("trip", at(1, 0, 0), at(5, 0, 0)),
		entry("last month", time.Date(2026, 9, 15, 10, 0, 0, 0, time.Local), time.Date(2026, 9, 15, 11, 0, 0, 0, time.Local)),
	}

	view := buildMonth(entries, parseMonth("2026-10", testNow), testNow, []string{"Private", "Family"})

	if view.Title != "October 2026" || view.Prev != "2026-09" || view.Next != "2026-11" || !view.IsCurrentMonth {
		t.Errorf("unexpected header: %q %q %q %v", view.Title, view.Prev, view.Next, view.IsCurrentMonth)
	}
	if len(view.Calendars) != 2 || view.Calendars[1].Name != "Family" || view.Calendars[1].Color != 1 {
		t.Errorf("unexpected legend: %+v", view.Calendars)
	}

	// october 2026 starts on a thursday and ends on a saturday
	if len(view.Weeks) != 5 {
		t.Fatalf("expected 5 weeks, got %d", len(view.Weeks))
	}
	first, last := view.Weeks[0][0], view.Weeks[4][6]
	if first.Key != "2026-09-28" || first.InMonth || last.Key != "2026-11-01" || last.InMonth {
		t.Errorf("unexpected grid: from %s to %s", first.Key, last.Key)
	}

	days := map[string]MonthDay{}
	for _, week := range view.Weeks {
		if len(week) != 7 {
			t.Fatalf("expected 7 days per week, got %d", len(week))
		}
		for _, day := range week {
			days[day.Key] = day
		}
	}

	today := days["2026-10-29"]
	if !today.Today || today.Past || today.Weekend || today.Label != "Today" {
		t.Errorf("unexpected flags for today: %+v", today)
	}
	if !days["2026-10-28"].Past || !days["2026-10-31"].Weekend {
		t.Error("expected the 28th to be past and the 31st to be a weekend")
	}

	if len(today.Entries) != 5 || len(today.Visible) != 2 || today.More != 3 {
		t.Fatalf("expected 5 entries, 2 visible and 3 more today, got %d, %d, %d", len(today.Entries), len(today.Visible), today.More)
	}
	if today.Entries[0].Summary != "offsite" || !today.Entries[0].Spans || today.Entries[0].When != "All day" {
		t.Errorf("expected the all-day entry first, got %+v", today.Entries[0])
	}
	if today.Entries[1].Summary != "run" || today.Entries[1].Spans || today.Entries[1].When != "07:00 – 08:00" {
		t.Errorf("unexpected second entry: %+v", today.Entries[1])
	}

	for _, key := range []string{"2026-10-30", "2026-10-31", "2026-11-01"} {
		if day := days[key]; len(day.Entries) != 1 || day.Entries[0].Summary != "trip" || !day.Entries[0].Spans {
			t.Errorf("%s: expected the trip, got %+v", key, day.Entries)
		}
	}
	if len(days["2026-10-28"].Entries) != 0 {
		t.Errorf("expected no entries on the 28th, got %+v", days["2026-10-28"].Entries)
	}

	// the list below the grid only contains days of the month with entries
	var listed []string
	for _, day := range view.Days {
		listed = append(listed, day.Key)
	}
	if strings.Join(listed, ",") != "2026-10-29,2026-10-30,2026-10-31" {
		t.Errorf("unexpected listed days: %v", listed)
	}

	september := buildMonth(entries, parseMonth("2026-09", testNow), testNow, nil)
	if september.IsCurrentMonth || len(september.Calendars) != 0 || len(september.Days) != 1 || september.Days[0].Entries[0].Summary != "last month" {
		t.Errorf("unexpected september: current=%v legend=%v days=%+v", september.IsCurrentMonth, september.Calendars, september.Days)
	}
}

func Test_parseMonth(t *testing.T) {
	if got := parseMonth("2027-02", testNow); got.Year() != 2027 || got.Month() != time.February || got.Day() != 1 {
		t.Errorf("unexpected month: %v", got)
	}
	for _, invalid := range []string{"", "2027", "february", "2027-13"} {
		if got := parseMonth(invalid, testNow); got.Year() != 2026 || got.Month() != time.October || got.Day() != 1 {
			t.Errorf("%q: expected the current month, got %v", invalid, got)
		}
	}
}

func TestCaldavDatasource_RenderPage(t *testing.T) {
	ds := &CaldavDatasource{location: time.Local}
	if ds.PageName() != "" {
		t.Error("expected no page without month template")
	}

	if err := WithMonthTemplate([]byte(`<h1>{{ .Title }}</h1>{{ range .Days }}{{ range .Entries }}<p>{{ .Summary }}</p>{{ end }}{{ end }}`))(ds); err != nil {
		t.Fatal(err)
	}
	if ds.PageName() != "calendar" {
		t.Errorf("unexpected page name %q", ds.PageName())
	}

	ds.entries = []Entry{entry("review", at(0, 9, 30), at(0, 10, 30))}
	page, err := ds.RenderPage(httptest.NewRequest("GET", "/calendar?month=2026-10", nil))
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "October 2026" || string(page.Content) != "<h1>October 2026</h1><p>review</p>" {
		t.Errorf("unexpected page: %q %q", page.Title, page.Content)
	}
}
