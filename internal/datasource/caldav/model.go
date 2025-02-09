package caldav

import (
	"fmt"
	"net/url"
	"time"

	"github.com/emersion/go-ical"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/pkg"
)

type CaldavData struct {
	Entries []Entry

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
	start       string
	Start       time.Time
	end         string
	End         time.Time
	Location    string
	LocationUrl string

	Formatted []string
}

func toEntry(event *ical.Component) Entry {
	entry := Entry{}

	prop := event.Props.Get("SUMMARY")
	if prop != nil {
		entry.Summary = prop.Value
	}

	prop = event.Props.Get("DTSTART")
	if prop != nil {
		entry.start = prop.Value
		entry.Start, _ = parseTime(entry.start)
	}

	prop = event.Props.Get("DTEND")
	if prop != nil {
		entry.end = prop.Value
		entry.End, _ = parseTime(entry.end)
	}

	prop = event.Props.Get("LOCATION")
	if prop != nil {
		entry.Location = prop.Value
		query := url.QueryEscape(entry.Location)
		entry.LocationUrl = fmt.Sprintf("https://google.com/maps/?q=%s", query)
	}

	if !entry.Start.IsZero() && !entry.End.IsZero() {
		entry.Formatted = getFormattedDate(entry.Start, entry.End)
	} else {
		log.Warn().Msg("can not properly format date")
		entry.Formatted = []string{fmt.Sprintf("%s – %s", entry.start, entry.end)}
	}

	return entry
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
			return []string{"Today", fmt.Sprintf("%s–%s", s, e)}
		}

		if pkg.IsTomorrow(start) {
			if isWholeDay {
				return []string{"Tomorrow"}
			}

			s := start.Format("15:04")
			e := end.Format("15:04")
			return []string{"Tomorrow", fmt.Sprintf("%s–%s", s, e)}
		}

		if isWholeDay {
			s := start.Format("02.01.")
			return []string{fmt.Sprintf("%s, %s", start.Weekday().String()[:3], s)}
		}

		w := start.Format("02.01.")
		s := start.Format("15:04")
		e := end.Format("15:04")
		return []string{fmt.Sprintf("%s, %s", start.Weekday().String()[:3], w), fmt.Sprintf("%s–%s", s, e)}
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

func parseTime(input string) (time.Time, error) {
	layoutTime := "20060102T150405"
	layoutDay := "20060102"

	// Parse the input string into a time.Time value
	parsedTime, err := time.Parse(layoutTime, input)
	if err != nil {
		parsedTime, err = time.Parse(layoutDay, input)
		if err != nil {
			log.Error().Err(err).Msg("caldav: error parsing time")
			return time.Time{}, err
		}
	}

	return parsedTime, nil
}
