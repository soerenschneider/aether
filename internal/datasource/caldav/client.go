package caldav

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"go.uber.org/multierr"
)

type ClientOpt func(client *Client) error

type Client struct {
	endpoint string
	username string
	password string
	// name or path of the calendar to fetch, the first calendar of the account if empty
	calendar string

	davClient  *caldav.Client
	httpClient *http.Client
}

func NewClient(endpoint string, opts ...ClientOpt) (*Client, error) {
	c := &Client{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(c); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	if errs != nil {
		return nil, errs
	}

	var client *caldav.Client
	var err error
	if len(c.username) > 0 && len(c.password) > 0 {
		client, err = caldav.NewClient(webdav.HTTPClientWithBasicAuth(c.httpClient, c.username, c.password), endpoint)
	} else {
		client, err = caldav.NewClient(c.httpClient, c.endpoint)
	}
	if err != nil {
		return nil, err
	}

	c.davClient = client
	return c, nil
}

func (c *Client) FetchData(ctx context.Context) ([]Entry, error) {
	homeSet, err := c.davClient.FindCalendarHomeSet(ctx, c.username)
	if err != nil {
		return nil, fmt.Errorf("finding home set: %w", err)
	}

	calendars, err := c.davClient.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("finding calendars: %w", err)
	}
	calendar, err := selectCalendar(calendars, c.calendar)
	if err != nil {
		return nil, err
	}

	//today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	resp, err := c.davClient.QueryCalendar(ctx, calendar.Path, &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			//Start: today,
			//End:   today.AddDate(0, 0, c.maxDays),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("querying calendar %q: %w", calendar.Path, err)
	}

	from, to := expansionWindow(time.Now())
	ignored := ignoredSummaries()

	var entries []Entry
	for _, object := range resp {
		for _, entry := range EntriesFromCalendar(object.Data, from, to, time.Local) {
			if !ignored[entry.Summary] {
				entries = append(entries, entry)
			}
		}
	}

	return entries, nil
}

// ignoredSummaries returns the summaries of events to leave out, configured as comma separated list in IGNORE.
func ignoredSummaries() map[string]bool {
	ret := map[string]bool{}
	for _, summary := range strings.Split(os.Getenv("IGNORE"), ",") {
		if summary = strings.TrimSpace(summary); summary != "" {
			ret[summary] = true
		}
	}
	return ret
}

// selectCalendar picks the calendar with the given name or path, or the first one if no calendar is wanted.
func selectCalendar(calendars []caldav.Calendar, wanted string) (caldav.Calendar, error) {
	if len(calendars) < 1 {
		return caldav.Calendar{}, errors.New("no calendars found")
	}
	if wanted == "" {
		return calendars[0], nil
	}

	for _, calendar := range calendars {
		if strings.EqualFold(calendar.Name, wanted) || calendar.Path == wanted || path.Base(calendar.Path) == strings.Trim(wanted, "/") {
			return calendar, nil
		}
	}

	available := make([]string, 0, len(calendars))
	for _, calendar := range calendars {
		available = append(available, fmt.Sprintf("%q (%s)", calendar.Name, calendar.Path))
	}
	return caldav.Calendar{}, fmt.Errorf("calendar %q not found, available: %s", wanted, strings.Join(available, ", "))
}
