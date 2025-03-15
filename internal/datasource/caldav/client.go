package caldav

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"go.uber.org/multierr"
)

type ClientOpt func(client *Client) error

type Client struct {
	endpoint string
	username string
	password string

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
	if len(calendars) < 1 {
		return nil, fmt.Errorf("no calendars found")
	}

	//today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	resp, err := c.davClient.QueryCalendar(ctx, calendars[0].Path, &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			//Start: today,
			//End:   today.AddDate(0, 0, c.maxDays),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("querying first calendar %q: %w", calendars[0].Path, err)
	}

	var entries []Entry
	for _, icsEvent := range resp {
		events := icsEvent.Data.Events()

		skipNames := strings.Split(os.Getenv("IGNORE"), ",")
		for _, e := range events {
			for _, igName := range skipNames {
				if evName := e.Props.Get("SUMMARY"); evName != nil &&
					evName.Value == igName {
					continue
				}
			}

			//redacted := redactComponent(e.Component)
			entry := toEntry(e.Component)
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
