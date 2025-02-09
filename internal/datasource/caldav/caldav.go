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
	"time"

	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"github.com/soerenschneider/aether/pkg"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"go.uber.org/multierr"
)

const (
	defaultMaxDays    = 7 * 4
	defaultMaxEntries = 10
)

type CaldavDatasource struct {
	endpoint string
	username string
	password string

	maxDays    int
	maxEntries int

	davClient *caldav.Client
	location  *time.Location

	defaultTemplate *template.Template
	simpleTemplate  *template.Template
	httpClient      *http.Client
}

type Opt func(datasource *CaldavDatasource) error

func New(endpoint string, templateData templates.TemplateData, opts ...Opt) (*CaldavDatasource, error) {
	if err := templateData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template data: %w", err)
	}

	ds := &CaldavDatasource{
		endpoint:   endpoint,
		maxDays:    defaultMaxDays,
		maxEntries: defaultMaxEntries,
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

	ds.defaultTemplate, err = template.New("agenda-regular").Funcs(template.FuncMap{
		"fixLocation": fixLocation,
	}).Parse(string(templateData.DefaultTemplate))
	if err != nil {
		return nil, err
	}

	if len(templateData.SimpleTemplate) > 0 {
		ds.simpleTemplate, err = template.New("agenda-simple").Funcs(template.FuncMap{
			"fixLocation": fixLocation,
		}).Parse(string(templateData.SimpleTemplate))
		if err != nil {
			return nil, err
		}
	}

	return ds, errs
}

func (c *CaldavDatasource) Name() string {
	return "Calendar"
}

func (c *CaldavDatasource) GetData(ctx context.Context) (*internal.Data, error) {
	data, err := c.getEntries(ctx)
	if err != nil {
		return nil, err
	}

	data.Entries = c.filter(data.Entries)
	sortEntries(data.Entries)
	if len(data.Entries) == c.maxEntries {
		data.To = data.Entries[len(data.Entries)-1].Start
	}

	data.HtmlId = pkg.NameToId(c.Name())

	var regularTemplateData bytes.Buffer
	if err := c.defaultTemplate.Execute(&regularTemplateData, data); err != nil {
		return nil, err
	}

	var simpleTemplateData bytes.Buffer
	if err := c.simpleTemplate.Execute(&simpleTemplateData, data); err != nil {
		return nil, err
	}

	return &internal.Data{
		Summary:                    getSummary(data.Entries),
		RenderedDefaultTemplate:    regularTemplateData.Bytes(),
		RenderedSimplifiedTemplate: simpleTemplateData.Bytes(),
	}, nil
}

func getSummary(entries []Entry) []string {
	now := time.Now()
	var ret []string
	for _, entry := range entries {
		isToday := pkg.IsToday(entry.Start, now)
		isOngoing := entry.Start.Before(now) && entry.End.After(now)

		if isToday || isOngoing {
			b := fmt.Sprintf("📅 %s, %s", entry.Summary, strings.Join(entry.Formatted, " "))
			ret = append(ret, b)
		}
	}

	return ret
}

func sortEntries(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Start.Before(entries[j].Start)
	})
}

func (c *CaldavDatasource) filter(entries []Entry) []Entry {
	var filtered []Entry
	start := pkg.Today(time.Now().In(c.location))
	end := pkg.NWeeks(time.Now().In(c.location), c.maxDays)

	for _, entry := range entries {
		if len(filtered) >= c.maxEntries {
			return filtered
		}

		if entry.Start.After(start) && entry.End.Before(end) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func fixLocation(input string) string {
	input = strings.Replace(input, "\\", "", -1)
	if !strings.HasSuffix(input, ")") {
		return input
	}

	lastIndex := strings.LastIndex(input, "(")
	if lastIndex > 0 {
		return input[0:lastIndex]
	}

	return input
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

	now := time.Now()
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

	daysUntilSunday := 7 - int(now.Weekday()) // Days until Sunday
	data := &CaldavData{
		Entries:         entries,
		From:            time.Now().In(c.location),
		To:              time.Now().In(c.location).AddDate(0, 0, c.maxDays),
		Now:             now,
		ThisWeekEnd:     now.AddDate(0, 0, daysUntilSunday),
		NextWeekEnd:     now.AddDate(0, 0, daysUntilSunday).AddDate(0, 0, 7),
		NextNextWeekEnd: now.AddDate(0, 0, daysUntilSunday).AddDate(0, 0, 14),
	}

	return data, nil
}
