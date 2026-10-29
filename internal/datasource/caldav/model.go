package caldav

import (
	"fmt"
	"net/url"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

type CaldavData struct {
	Entries []Entry
	// names of the calendars that could not be fetched
	Failed []string
	// link to the month view, empty if there is none
	MonthLink string

	HtmlId string
	From   time.Time
	To     time.Time

	Now             time.Time
	ThisWeekEnd     time.Time
	NextWeekEnd     time.Time
	NextNextWeekEnd time.Time
}

type Entry struct {
	Summary     string
	Start       time.Time
	End         time.Time
	Location    string
	LocationUrl string

	// name of the calendar(s) the entry belongs to, only set if multiple calendars are configured
	Calendar string
	// index into a palette, to tell calendars apart
	CalendarColor int

	Formatted []string
}

// NewEntry builds an entry from already parsed values, e.g. for fake data during development.
func NewEntry(summary, location string, start, end time.Time) Entry {
	entry := Entry{
		Summary:   summary,
		Start:     start,
		End:       end,
		Location:  location,
		Formatted: getFormattedDate(start, end),
	}
	if location != "" {
		entry.LocationUrl = buildLocationUrl(location)
	}
	return entry
}

func buildLocationUrl(location string) string {
	return fmt.Sprintf("https://google.com/maps/?q=%s", url.QueryEscape(location))
}

func getFormattedDate(start, end time.Time) []string {
	now := time.Now()
	isWholeDay := pkg.IsWholeDay(start, end)

	if pkg.AtSameDay(start, end) {
		if pkg.IsToday(start, now) {
			if isWholeDay {
				return []string{"Today"}
			}

			s := start.Format("15:04")
			e := end.Format("15:04")
			return []string{"Today", fmt.Sprintf("%s – %s", s, e)}
		}

		if pkg.IsTomorrow(start) {
			if isWholeDay {
				return []string{"Tomorrow"}
			}

			s := start.Format("15:04")
			e := end.Format("15:04")
			return []string{"Tomorrow", fmt.Sprintf("%s – %s", s, e)}
		}

		if isWholeDay {
			s := start.Format("02.01.")
			return []string{fmt.Sprintf("%s, %s", start.Weekday().String()[:3], s)}
		}

		w := start.Format("02.01.")
		s := start.Format("15:04")
		e := end.Format("15:04")
		return []string{fmt.Sprintf("%s, %s", start.Weekday().String()[:3], w), fmt.Sprintf("%s – %s", s, e)}
	}

	dur := end.Sub(start)
	if int(dur.Hours()/24) > 0 {
		days := int64(dur.Hours()) / 24
		if int64(dur.Hours())%24 > 0 {
			days += 1
		}
		return []string{fmt.Sprintf("%s – %s (%d days)", formatDate(start, isWholeDay), formatDate(end, isWholeDay), days)}
	}

	return []string{fmt.Sprintf("%s – %s", formatDate(start, isWholeDay), formatDate(end, isWholeDay))}
}

func formatDate(date time.Time, isWholeDay bool) string {
	format := "02.01. 15:04"
	if isWholeDay {
		format = "02.01."
	}
	return fmt.Sprintf("%s, %s", date.Weekday().String()[:3], date.Format(format))
}
