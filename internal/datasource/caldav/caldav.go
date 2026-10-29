package caldav

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"github.com/soerenschneider/aether/pkg"
)

const (
	defaultMaxDays    = 7 * 4
	defaultMaxEntries = 10
)

type CaldavDatasource struct {
	client CaldavClient

	maxDays    int
	maxEntries int

	location *time.Location

	defaultTemplate    *template.Template
	simpleTemplate     *template.Template
	monthTemplate      *template.Template
	excludeFromSummary bool

	// all entries of the last update, sorted, for the month view
	mutex   sync.RWMutex
	entries []Entry
	failed  []string
}

type CaldavClient interface {
	FetchData(ctx context.Context) ([]Entry, error)
}

type DatasourceOpt func(datasource *CaldavDatasource) error

func New(client CaldavClient, templateData templates.TemplateData, opts ...DatasourceOpt) (*CaldavDatasource, error) {
	if err := templateData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template data: %w", err)
	}

	if client == nil {
		return nil, errors.New("nil client passed")
	}

	ds := &CaldavDatasource{
		client:     client,
		maxDays:    defaultMaxDays,
		maxEntries: defaultMaxEntries,
		location:   time.Now().Location(),
	}

	for _, opt := range opts {
		if err := opt(ds); err != nil {
			return nil, err
		}
	}

	var err error
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

	return ds, nil
}

func (c *CaldavDatasource) Name() string {
	return "Calendar"
}

func (c *CaldavDatasource) GetData(ctx context.Context) (*internal.Data, error) {
	entries, err := c.client.FetchData(ctx)
	var partialErr *PartialError
	if errors.As(err, &partialErr) {
		// show the calendars that could be fetched
		log.Warn().Err(err).Msg("could not fetch all calendars")
	} else if err != nil {
		return nil, err
	}

	now := time.Now()
	daysUntilSunday := 7 - int(now.Weekday()) // Days until Sunday
	data := CaldavData{
		Entries:         entries,
		From:            time.Now().In(c.location),
		To:              time.Now().In(c.location).AddDate(0, 0, c.maxDays),
		Now:             now,
		ThisWeekEnd:     now.AddDate(0, 0, daysUntilSunday),
		NextWeekEnd:     now.AddDate(0, 0, daysUntilSunday).AddDate(0, 0, 7),
		NextNextWeekEnd: now.AddDate(0, 0, daysUntilSunday).AddDate(0, 0, 14),
	}

	if partialErr != nil {
		data.Failed = partialErr.Calendars
	}
	data.MonthLink = c.PageName()

	// sort before filtering, as the filter stops after the maximum amount of entries
	sortEntries(data.Entries)
	c.mutex.Lock()
	c.entries = data.Entries
	c.failed = data.Failed
	c.mutex.Unlock()
	data.Entries = c.filter(data.Entries)
	if len(data.Entries) == c.maxEntries {
		data.To = data.Entries[len(data.Entries)-1].Start
	}

	data.HtmlId = pkg.NameToId(c.Name())

	var regularTemplateData bytes.Buffer
	if err := c.defaultTemplate.Execute(&regularTemplateData, data); err != nil {
		return nil, fmt.Errorf("could not render 'regular' template: %w: %w", internal.ErrTemplate, err)
	}

	var simpleTemplateData bytes.Buffer
	if err := c.simpleTemplate.Execute(&simpleTemplateData, data); err != nil {
		return nil, fmt.Errorf("could not render 'simple' template: %w: %w", internal.ErrTemplate, err)
	}

	var summary []string
	if !c.excludeFromSummary {
		summary = getSummary(data.Entries, time.Now(), true)
	}

	tile, tileErr := internal.RenderTile(c.defaultTemplate, data)
	if tileErr != nil {
		return nil, fmt.Errorf("could not render tile for datasource %q: %w", c.Name(), tileErr)
	}

	return &internal.Data{
		Summary:                    summary,
		RenderedDefaultTemplate:    regularTemplateData.Bytes(),
		RenderedSimplifiedTemplate: simpleTemplateData.Bytes(),
		RenderedTile:               tile,
		Incomplete:                 partialErr != nil,
	}, nil
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

		// keep everything that has not ended before today, e.g. all-day events starting at midnight or multi-day
		// events that started earlier
		if entry.End.After(start) && entry.Start.Before(end) {
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
