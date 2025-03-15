package caldav

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

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
	excludeFromSummary bool
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
	if err != nil {
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

	data.Entries = c.filter(data.Entries)
	sortEntries(data.Entries)
	if len(data.Entries) == c.maxEntries {
		data.To = data.Entries[len(data.Entries)-1].Start
	}

	data.HtmlId = pkg.NameToId(c.Name())

	var regularTemplateData bytes.Buffer
	if err := c.defaultTemplate.Execute(&regularTemplateData, data); err != nil {
		return nil, fmt.Errorf("could not render 'regular' template: %w", internal.ErrTemplate)
	}

	var simpleTemplateData bytes.Buffer
	if err := c.simpleTemplate.Execute(&simpleTemplateData, data); err != nil {
		return nil, fmt.Errorf("could not render 'simple' template: %w", internal.ErrTemplate)
	}

	var summary []string
	if !c.excludeFromSummary {
		summary = getSummary(data.Entries, time.Now(), true)
	}

	return &internal.Data{
		Summary:                    summary,
		RenderedDefaultTemplate:    regularTemplateData.Bytes(),
		RenderedSimplifiedTemplate: simpleTemplateData.Bytes(),
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
