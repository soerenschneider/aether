package caldav

import (
	"fmt"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

// helpers that prepare the entries for the templates, all relative to CaldavData.Now

type AgendaDay struct {
	// Today, Tomorrow or the date
	Label   string
	Entries []AgendaEntry
}

type AgendaEntry struct {
	Entry
	// time of the entry within its day, e.g. "10:00 – 10:30", "All day" or "until Fri 18:00"
	When    string
	Past    bool
	Ongoing bool
}

func (e Entry) isAllDay() bool {
	return pkg.IsWholeDay(e.Start, e.End)
}

// Days groups the entries by day. Entries that started before today but have not ended yet are shown today.
func (d CaldavData) Days() []AgendaDay {
	today := pkg.Today(d.Now)

	var days []AgendaDay
	var currentDay time.Time
	for _, entry := range d.Entries {
		day := pkg.Today(entry.Start)
		if day.Before(today) {
			day = today
		}

		if len(days) == 0 || !day.Equal(currentDay) {
			currentDay = day
			days = append(days, AgendaDay{Label: d.dayLabel(day)})
		}

		agendaEntry := AgendaEntry{
			Entry: entry,
			When:  d.entryTime(entry),
		}
		if !entry.isAllDay() {
			agendaEntry.Past = !entry.End.After(d.Now)
			agendaEntry.Ongoing = !entry.Start.After(d.Now) && entry.End.After(d.Now)
		}

		days[len(days)-1].Entries = append(days[len(days)-1].Entries, agendaEntry)
	}

	return days
}

// Current returns the timed entry that is running right now, if any.
func (d CaldavData) Current() *Entry {
	for i, entry := range d.Entries {
		if !entry.isAllDay() && !entry.Start.After(d.Now) && entry.End.After(d.Now) {
			return &d.Entries[i]
		}
	}
	return nil
}

// Next returns the first entry that starts in the future.
func (d CaldavData) Next() *Entry {
	for i, entry := range d.Entries {
		if entry.Start.After(d.Now) {
			return &d.Entries[i]
		}
	}
	return nil
}

// LaterToday counts the entries that start later today.
func (d CaldavData) LaterToday() int {
	count := 0
	for _, entry := range d.Entries {
		if entry.Start.After(d.Now) && pkg.IsToday(entry.Start, d.Now) {
			count++
		}
	}
	return count
}

// MoreToday counts the entries that start later today, besides the one returned by Next.
func (d CaldavData) MoreToday() int {
	return max(0, d.LaterToday()-1)
}

// ShortWhen describes the start of an entry as briefly as possible, e.g. "12:00", "tomorrow 19:00" or "Mon 02.11.".
func (d CaldavData) ShortWhen(entry *Entry) string {
	if entry == nil {
		return ""
	}

	day := pkg.Today(entry.Start)
	label := d.dayLabel(day)
	if entry.isAllDay() {
		return label
	}

	if day.Equal(pkg.Today(d.Now)) {
		return entry.Start.Format("15:04")
	}
	if label == "Tomorrow" {
		return "tomorrow " + entry.Start.Format("15:04")
	}
	return fmt.Sprintf("%s %s", entry.Start.Format("Mon"), entry.Start.Format("15:04"))
}

func (d CaldavData) dayLabel(day time.Time) string {
	today := pkg.Today(d.Now)
	switch {
	case day.Equal(today):
		return "Today"
	case day.Equal(today.AddDate(0, 0, 1)):
		return "Tomorrow"
	default:
		return day.Format("Mon 02.01.")
	}
}

func (d CaldavData) entryTime(entry Entry) string {
	if pkg.AtSameDay(entry.Start, entry.End) {
		if entry.isAllDay() {
			return "All day"
		}
		return fmt.Sprintf("%s – %s", entry.Start.Format("15:04"), entry.End.Format("15:04"))
	}

	// multi-day entries
	if entry.isAllDay() {
		// the end of all-day entries is exclusive
		return "until " + d.dayLabelShort(entry.End.Add(-time.Second))
	}
	return fmt.Sprintf("until %s %s", d.dayLabelShort(entry.End), entry.End.Format("15:04"))
}

func (d CaldavData) dayLabelShort(t time.Time) string {
	day := pkg.Today(t)
	switch {
	case day.Equal(pkg.Today(d.Now)):
		return "today"
	case day.Equal(pkg.Today(d.Now).AddDate(0, 0, 1)):
		return "tomorrow"
	default:
		return day.Format("Mon 02.01.")
	}
}
