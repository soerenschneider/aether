package caldav

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/rs/zerolog/log"
	"github.com/teambition/rrule-go"
)

// converting calendar objects to entries, including the expansion of recurring events

const (
	// maximum amount of occurrences of a single recurring event within the window, protects against rules like
	// FREQ=MINUTELY
	maxOccurrences = 1000
	// maximum amount of iterations to reach the window, protects against old rules with a high frequency
	maxIterations = 100_000
)

// expansionWindow returns the time span recurring events are expanded for. The month view can show past and future
// months, so it's a year in both directions.
func expansionWindow(now time.Time) (time.Time, time.Time) {
	return now.AddDate(-1, 0, 0), now.AddDate(1, 0, 0)
}

// EntriesFromCalendar converts the events of a calendar object to entries in the given location. Recurring events
// are expanded to their occurrences between from and to, honoring EXDATE, RDATE and modified occurrences
// (RECURRENCE-ID). Cancelled events and occurrences are left out.
func EntriesFromCalendar(cal *ical.Calendar, from, to time.Time, loc *time.Location) []Entry {
	events := cal.Events()

	// occurrences that are replaced by a modified (or cancelled) instance, by UID and original start
	replaced := map[string]map[int64]bool{}
	for _, event := range events {
		prop := event.Props.Get(ical.PropRecurrenceID)
		if prop == nil {
			continue
		}
		recurrenceID, err := prop.DateTime(loc)
		if err != nil {
			log.Warn().Err(err).Msg("caldav: could not parse RECURRENCE-ID")
			continue
		}
		uid := propText(event, ical.PropUID)
		if replaced[uid] == nil {
			replaced[uid] = map[int64]bool{}
		}
		replaced[uid][recurrenceID.Unix()] = true
	}

	var entries []Entry
	for _, event := range events {
		if status, _ := event.Status(); status == ical.EventCancelled {
			continue
		}

		start, err := event.DateTimeStart(loc)
		if err != nil || start.IsZero() {
			log.Warn().Err(err).Str("summary", propText(event, ical.PropSummary)).Msg("caldav: event without valid start")
			continue
		}
		end, err := event.DateTimeEnd(loc)
		if err != nil || end.Before(start) {
			end = start
		}

		summary := propText(event, ical.PropSummary)
		location := propText(event, ical.PropLocation)
		single := NewEntry(summary, location, start.In(loc), end.In(loc))

		// modified occurrences are regular events
		if event.Props.Get(ical.PropRecurrenceID) != nil {
			entries = append(entries, single)
			continue
		}

		set, err := recurrenceSet(event, start, loc)
		if err != nil {
			log.Warn().Err(err).Str("summary", summary).Msg("caldav: could not parse recurrence, showing first occurrence only")
		}
		if set == nil {
			entries = append(entries, single)
			continue
		}

		allDay := isDateValue(event.Props.Get(ical.PropDateTimeStart))
		days := int(math.Round(end.Sub(start).Hours() / 24))
		duration := end.Sub(start)

		for _, occurrence := range occurrences(set, from, to) {
			if replaced[propText(event, ical.PropUID)][occurrence.Unix()] {
				continue
			}

			// all-day events keep their calendar days, also across daylight saving time changes
			occurrenceEnd := occurrence.Add(duration)
			if allDay {
				occurrenceEnd = occurrence.AddDate(0, 0, days)
			}
			entries = append(entries, NewEntry(summary, location, occurrence.In(loc), occurrenceEnd.In(loc)))
		}
	}

	return entries
}

// occurrences returns the starts of the occurrences that take place between from and to.
func occurrences(set *rrule.Set, from, to time.Time) []time.Time {
	var ret []time.Time
	next := set.Iterator()
	for i := 0; i < maxIterations && len(ret) < maxOccurrences; i++ {
		occurrence, ok := next()
		if !ok || occurrence.After(to) {
			break
		}
		if !occurrence.Before(from) {
			ret = append(ret, occurrence)
		}
	}
	return ret
}

// recurrenceSet builds the recurrence set of an event, nil if the event does not recur. Unlike
// ical.Component.RecurrenceSet, it supports RDATE as well as lists of dates in EXDATE and RDATE.
func recurrenceSet(event ical.Event, start time.Time, loc *time.Location) (*rrule.Set, error) {
	option, err := event.Props.RecurrenceRule()
	if err != nil {
		return nil, err
	}
	if option == nil {
		return nil, nil
	}

	// occurrences are calculated in the time zone of the start, so they keep their local time across daylight
	// saving time changes
	option.Dtstart = start
	rule, err := rrule.NewRRule(*option)
	if err != nil {
		return nil, fmt.Errorf("invalid recurrence rule: %w", err)
	}

	set := &rrule.Set{}
	set.RRule(rule)

	exdates, err := dateList(event.Props[ical.PropExceptionDates], loc)
	if err != nil {
		return nil, fmt.Errorf("invalid EXDATE: %w", err)
	}
	set.SetExDates(exdates)

	rdates, err := dateList(event.Props[ical.PropRecurrenceDates], loc)
	if err != nil {
		return nil, fmt.Errorf("invalid RDATE: %w", err)
	}
	for _, rdate := range rdates {
		set.RDate(rdate)
	}

	return set, nil
}

// dateList parses date properties, each of which may contain a comma separated list of dates.
func dateList(props []ical.Prop, loc *time.Location) ([]time.Time, error) {
	var ret []time.Time
	for _, prop := range props {
		for _, value := range strings.Split(prop.Value, ",") {
			single := prop
			single.Value = strings.TrimSpace(value)
			// periods ("start/end" or "start/duration") are only supported by their start
			single.Value, _, _ = strings.Cut(single.Value, "/")
			if single.Value == "" {
				continue
			}

			date, err := single.DateTime(loc)
			if err != nil {
				return nil, err
			}
			ret = append(ret, date)
		}
	}
	return ret, nil
}

func isDateValue(prop *ical.Prop) bool {
	if prop == nil {
		return false
	}
	return prop.ValueType() == ical.ValueDate || (prop.ValueType() == ical.ValueDefault && len(prop.Value) == len("20060102"))
}

func propText(event ical.Event, name string) string {
	prop := event.Props.Get(name)
	if prop == nil {
		return ""
	}
	text, err := prop.Text()
	if err != nil {
		return prop.Value
	}
	return text
}
