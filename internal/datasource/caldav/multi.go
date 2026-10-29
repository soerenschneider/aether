package caldav

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// number of distinct colors the templates use to tell calendars apart
const calendarColors = 5

type NamedCalendar struct {
	Name   string
	Client CaldavClient
}

// MultiClient fetches several calendars and merges their entries.
type MultiClient struct {
	calendars []NamedCalendar
}

// PartialError is returned by MultiClient if some, but not all, calendars could not be fetched. The entries of the
// other calendars are returned along with it.
type PartialError struct {
	Calendars []string
	Err       error
}

func (e *PartialError) Error() string {
	return fmt.Sprintf("could not fetch calendars %s: %v", strings.Join(e.Calendars, ", "), e.Err)
}

func (e *PartialError) Unwrap() error {
	return e.Err
}

func NewMultiClient(calendars ...NamedCalendar) (*MultiClient, error) {
	if len(calendars) == 0 {
		return nil, errors.New("no calendars passed")
	}

	for _, calendar := range calendars {
		if calendar.Client == nil {
			return nil, fmt.Errorf("nil client passed for calendar %q", calendar.Name)
		}
	}

	return &MultiClient{calendars: calendars}, nil
}

func (m *MultiClient) FetchData(ctx context.Context) ([]Entry, error) {
	results := make([][]Entry, len(m.calendars))
	errs := make([]error, len(m.calendars))

	var wg sync.WaitGroup
	for i, calendar := range m.calendars {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = calendar.Client.FetchData(ctx)
		}()
	}
	wg.Wait()

	var failed []string
	var entries []Entry
	// only label entries if there's more than one calendar to tell apart
	labelled := len(m.calendars) > 1
	// the same event can show up in several calendars, e.g. a family event one was invited to
	seen := map[string]int{}

	for i, calendar := range m.calendars {
		if errs[i] != nil {
			failed = append(failed, calendar.Name)
			continue
		}

		for _, entry := range results[i] {
			if labelled {
				entry.Calendar = calendar.Name
				entry.CalendarColor = i % calendarColors
			}

			key := fmt.Sprintf("%s|%d|%d", entry.Summary, entry.Start.Unix(), entry.End.Unix())
			if idx, ok := seen[key]; ok {
				if labelled {
					entries[idx].Calendar += ", " + calendar.Name
				}
				continue
			}

			seen[key] = len(entries)
			entries = append(entries, entry)
		}
	}

	if len(failed) == len(m.calendars) {
		return nil, errors.Join(errs...)
	}
	if len(failed) > 0 {
		return entries, &PartialError{Calendars: failed, Err: errors.Join(errs...)}
	}
	return entries, nil
}
