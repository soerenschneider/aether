package caldav

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/soerenschneider/aether/pkg"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"go.uber.org/multierr"
)

const defaultDays = 7 * 4

type CaldavDatasource struct {
	endpoint string
	username string
	password string

	days int

	davClient *caldav.Client
	location  *time.Location

	template   *template.Template
	once       sync.Once
	httpClient *http.Client
}

type Opt func(datasource *CaldavDatasource) error

func New(endpoint string, opts ...Opt) (*CaldavDatasource, error) {
	ds := &CaldavDatasource{
		endpoint:   endpoint,
		days:       defaultDays,
		httpClient: http.DefaultClient,
		location:   time.Now().Location(),
	}

	var errs error
	for _, opt := range opts {
		if err := opt(ds); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	var client *caldav.Client
	var err error
	if len(ds.username) > 0 && len(ds.password) > 0 {
		client, err = caldav.NewClient(webdav.HTTPClientWithBasicAuth(ds.httpClient, ds.username, ds.password), endpoint)
	} else {
		client, err = caldav.NewClient(ds.httpClient, ds.endpoint)
	}

	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("creating client: %w", err))
		return nil, errs
	}

	ds.davClient = client
	return ds, errs
}

func (c *CaldavDatasource) Name() string {
	return "Caldav"
}

func (c *CaldavDatasource) GetHtml(ctx context.Context) (string, error) {
	c.once.Do(func() {
		if c.template == nil {
			c.template = template.Must(template.New("agenda").Parse(defaultTemplate))
		}
	})

	data, err := c.getEntries(ctx)
	if err != nil {
		return "", err
	}
	data.Entries = c.filter(data.Entries)
	sortEntries(data.Entries)

	var tpl bytes.Buffer
	err = c.template.Execute(&tpl, data)
	return tpl.String(), err
}

func sortEntries(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Start.Before(entries[j].Start)
	})
}

func (c *CaldavDatasource) filter(entries []Entry) []Entry {
	var filtered []Entry
	start := pkg.Today(time.Now().In(c.location))
	end := pkg.NWeeks(time.Now().In(c.location), c.days)

	for _, entry := range entries {
		if entry.Start.After(start) && entry.End.Before(end) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func (c *CaldavDatasource) getEntries(ctx context.Context) (*CaldavData, error) {
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

	resp, err := c.davClient.QueryCalendar(ctx, calendars[0].Path, &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
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

	data := &CaldavData{
		Entries: entries,
		From:    time.Now().In(c.location),
		To:      time.Now().In(c.location).AddDate(0, 0, c.days),
	}

	return data, nil
}
