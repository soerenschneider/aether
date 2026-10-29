package caldav

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/pkg"
)

// a month view of the calendar, offered as a page next to the dashboard

const (
	monthPageName = "calendar"
	// entries shown per day in the grid, the rest is summarized as "+N more"
	visibleEntriesPerDay = 3
)

type MonthView struct {
	// e.g. "October 2026"
	Title string
	// months to navigate to, formatted as "2006-01"
	Prev           string
	Next           string
	IsCurrentMonth bool

	Weekdays []string
	Weeks    [][]MonthDay
	// the days of the month that have entries, listed below the grid on small screens
	Days []MonthDay

	// only set if there's more than one calendar
	Calendars []CalendarLegend
	Failed    []string
}

type CalendarLegend struct {
	Name  string
	Color int
}

type MonthDay struct {
	Date time.Time
	// e.g. "2026-10-29", used as anchor
	Key string
	// e.g. "Today" or "Thu 29.10."
	Label   string
	InMonth bool
	Today   bool
	Past    bool
	Weekend bool

	Entries []MonthEntry
	// the entries that fit into the grid
	Visible []MonthEntry
	More    int
}

type MonthEntry struct {
	Entry
	// all-day and multi-day entries are shown as bars
	Spans bool
	// time of the entry within the day, e.g. "10:00 – 10:30" or "All day"
	When string
}

// parseMonth parses a month formatted as "2006-01", the current month is returned for invalid input.
func parseMonth(value string, now time.Time) time.Time {
	if month, err := time.ParseInLocation("2006-01", value, now.Location()); err == nil {
		return month
	}
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

func buildMonth(entries []Entry, month, now time.Time, calendars []string) MonthView {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, now.Location())
	last := first.AddDate(0, 1, -1)
	today := pkg.Today(now)

	view := MonthView{
		Title:          first.Format("January 2006"),
		Prev:           first.AddDate(0, -1, 0).Format("2006-01"),
		Next:           first.AddDate(0, 1, 0).Format("2006-01"),
		IsCurrentMonth: first.Year() == now.Year() && first.Month() == now.Month(),
		Weekdays:       []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
	}

	if len(calendars) > 1 {
		for i, name := range calendars {
			view.Calendars = append(view.Calendars, CalendarLegend{Name: name, Color: i % calendarColors})
		}
	}

	// weeks start on monday
	start := first.AddDate(0, 0, -((int(first.Weekday()) + 6) % 7))
	end := last.AddDate(0, 0, (7-int(last.Weekday()))%7)

	labeler := CaldavData{Now: now}
	var week []MonthDay
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		day := MonthDay{
			Date:    date,
			Key:     date.Format("2006-01-02"),
			Label:   labeler.dayLabel(date),
			InMonth: date.Month() == first.Month(),
			Today:   date.Equal(today),
			Past:    date.Before(today),
			Weekend: date.Weekday() == time.Saturday || date.Weekday() == time.Sunday,
		}
		day.Entries = entriesOfDay(entries, date, labeler)
		day.Visible = day.Entries
		if len(day.Entries) > visibleEntriesPerDay {
			// keep the last line for the "+N more" hint
			day.Visible = day.Entries[:visibleEntriesPerDay-1]
			day.More = len(day.Entries) - len(day.Visible)
		}

		week = append(week, day)
		if len(week) == 7 {
			view.Weeks = append(view.Weeks, week)
			week = nil
		}
		if day.InMonth && len(day.Entries) > 0 {
			view.Days = append(view.Days, day)
		}
	}

	return view
}

// entriesOfDay returns the entries that take place on the given day, all-day and multi-day entries first.
func entriesOfDay(entries []Entry, date time.Time, labeler CaldavData) []MonthEntry {
	dayEnd := date.AddDate(0, 0, 1)

	var ret []MonthEntry
	for _, entry := range entries {
		end := entry.End
		if !end.After(entry.Start) {
			// entries without duration still take place
			end = entry.Start.Add(time.Minute)
		}
		if !entry.Start.Before(dayEnd) || !end.After(date) {
			continue
		}

		ret = append(ret, MonthEntry{
			Entry: entry,
			Spans: entry.isAllDay() || !pkg.AtSameDay(entry.Start, entry.End),
			When:  labeler.entryTime(entry),
		})
	}

	sort.SliceStable(ret, func(i, j int) bool {
		if ret[i].Spans != ret[j].Spans {
			return ret[i].Spans
		}
		return ret[i].Start.Before(ret[j].Start)
	})
	return ret
}

// PageName implements internal.PageProvider, the month view is only offered if a template for it is configured.
func (c *CaldavDatasource) PageName() string {
	if c.monthTemplate == nil {
		return ""
	}
	return monthPageName
}

// RenderPage renders the month given by the "month" query parameter, formatted as "2006-01". It uses the entries
// of the last update, so it does not query the calendars.
func (c *CaldavDatasource) RenderPage(r *http.Request) (*internal.Page, error) {
	if c.monthTemplate == nil {
		return nil, fmt.Errorf("no month template configured")
	}

	c.mutex.RLock()
	entries := c.entries
	failed := c.failed
	c.mutex.RUnlock()

	now := time.Now().In(c.location)
	view := buildMonth(entries, parseMonth(r.URL.Query().Get("month"), now), now, c.calendarNames())
	view.Failed = failed

	var buf bytes.Buffer
	if err := c.monthTemplate.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("could not render month template: %w", err)
	}

	return &internal.Page{
		Title:   view.Title,
		Content: buf.Bytes(),
	}, nil
}

func (c *CaldavDatasource) calendarNames() []string {
	if named, ok := c.client.(interface{ Calendars() []string }); ok {
		return named.Calendars()
	}
	return nil
}
