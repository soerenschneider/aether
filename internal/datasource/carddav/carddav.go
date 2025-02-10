package carddav

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/carddav"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"github.com/soerenschneider/aether/pkg"
	"go.uber.org/multierr"
)

const defaultLookaheadDays = 7 * 2

type Opt func(datasource *CarddavDatasource) error

type CarddavDatasource struct {
	endpoint      string
	username      string
	password      string
	lookaheadDays int

	davClient          *carddav.Client
	regularTemplate    *template.Template
	simpleTemplate     *template.Template
	httpClient         *http.Client
	excludeFromSummary bool
}

func New(endpoint string, templateData templates.TemplateData, opts ...Opt) (*CarddavDatasource, error) {
	if err := templateData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template data: %w", err)
	}

	ds := &CarddavDatasource{
		endpoint:      endpoint,
		lookaheadDays: defaultLookaheadDays,
		httpClient:    http.DefaultClient,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(ds); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	var client *carddav.Client
	var err error
	if len(ds.username) > 0 && len(ds.password) > 0 {
		client, err = carddav.NewClient(webdav.HTTPClientWithBasicAuth(ds.httpClient, ds.username, ds.password), endpoint)
	} else {
		client, err = carddav.NewClient(ds.httpClient, ds.endpoint)
	}

	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("creating client: %w", err))
		return nil, errs
	}

	ds.davClient = client

	ds.regularTemplate, err = template.New("anniversaries-regular").Parse(string(templateData.DefaultTemplate))
	if err != nil {
		return nil, err
	}

	if len(templateData.SimpleTemplate) > 0 {
		ds.simpleTemplate, err = template.New("anniversaries-simple").Parse(string(templateData.SimpleTemplate))
		if err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (c *CarddavDatasource) Name() string {
	return "Contacts"
}

func (c *CarddavDatasource) GetData(ctx context.Context) (*internal.Data, error) {
	data, err := c.getEntries(ctx)
	if err != nil {
		return nil, err
	}

	data.Cards = c.filter(data.Cards)
	sortCards(data.Cards, time.Now())

	var regularTemplateData bytes.Buffer
	if err := c.regularTemplate.Execute(&regularTemplateData, data); err != nil {
		return nil, err
	}

	var simpleTemplateData bytes.Buffer
	if err := c.simpleTemplate.Execute(&simpleTemplateData, data); err != nil {
		return nil, err
	}

	var summary []string
	if !c.excludeFromSummary {
		summary = getSummary(data.Cards, time.Now(), true)
	}

	return &internal.Data{
		Summary:                    summary,
		RenderedDefaultTemplate:    regularTemplateData.Bytes(),
		RenderedSimplifiedTemplate: simpleTemplateData.Bytes(),
	}, nil
}

func getSummary(entries []Card, now time.Time) []string {
	var ret []string
	for _, entry := range entries {
		anniversary := time.Date(now.Year(), entry.Anniversary.Month(), entry.Anniversary.Day(), 12, 0, 0, 0, time.UTC)
		if pkg.IsToday(anniversary, now) {
			b := fmt.Sprintf("%s, %s", entry.Name, entry.Type)
			ret = append(ret, b)
		}
	}

	return ret
}

func sortCards(entries []Card, now time.Time) {
	sort.Slice(entries, func(i, j int) bool {
		iMonth := entries[i].Anniversary.Month()
		jMonth := entries[j].Anniversary.Month()
		curMonth := now.Month()

		if iMonth >= curMonth && jMonth >= curMonth || iMonth < curMonth && jMonth < curMonth {
			if iMonth < jMonth {
				return true
			}
			if iMonth > jMonth {
				return false
			}

			return entries[i].Anniversary.Day() < entries[j].Anniversary.Day()
		}

		if iMonth < curMonth && jMonth >= curMonth || iMonth >= curMonth && jMonth < curMonth {
			return true
		}

		return entries[i].Anniversary.Day() < entries[j].Anniversary.Day()
	})
}

func (c *CarddavDatasource) isUpcoming(anniversary time.Time) bool {
	currentDate := time.Now()
	compareYear := currentDate.Year()
	if anniversary.Month() < currentDate.Month() {
		compareYear += 1
	}
	anniversary = time.Date(compareYear, anniversary.Month(), anniversary.Day(), 0, 0, 0, 0, currentDate.Location())
	currentDate = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())

	difference := anniversary.Sub(currentDate)
	return difference >= 0 && difference <= time.Duration(c.lookaheadDays)*24*time.Hour
}

func (c *CarddavDatasource) filter(entries []Card) []Card {
	var filtered []Card

	for _, entry := range entries {
		if !entry.Anniversary.IsZero() && c.isUpcoming(entry.Anniversary) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func (c *CarddavDatasource) getEntries(ctx context.Context) (*CarddavData, error) {
	homeSet, err := c.davClient.FindAddressBookHomeSet(ctx, c.username)
	if err != nil {
		return nil, fmt.Errorf("carddav: could not find homeset %w", err)
	}

	addressbooks, err := c.davClient.FindAddressBooks(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("carddav: could not find addressbooks: %w", err)
	}

	if len(addressbooks) < 1 {
		return nil, fmt.Errorf("no addressbooks found")
	}

	var entries []Card
	q := carddav.AddressBookQuery{
		DataRequest: carddav.AddressDataRequest{
			Props: []string{
				vcard.FieldBirthday,
				vcard.FieldAnniversary,
				vcard.FieldName,
			},
		},
	}

	resp, err := c.davClient.QueryAddressBook(ctx, addressbooks[0].Path, &q)
	if err != nil {
		return nil, fmt.Errorf("querying first calendar %q: %w", addressbooks[0].Path, err)
	}

	now := time.Now()
	for _, r := range resp {
		birthday := r.Card.Get(vcard.FieldBirthday)
		if birthday != nil {
			card, err := buildCard(r.Card, birthday, "Birthday 🎂", now)
			if err != nil {
				log.Error().Err(err).Msg("could not extract anniversary")
			} else {
				entries = append(entries, card)
			}
		}

		anniversary := r.Card.Get(vcard.FieldAnniversary)
		if anniversary != nil {
			card, err := buildCard(r.Card, anniversary, "Anniversary 🥂", now)
			if err != nil {
				log.Error().Err(err).Msg("could not extract anniversary")
			} else {
				entries = append(entries, card)
			}
		}
	}

	return &CarddavData{
		Cards:  entries,
		From:   time.Now(),
		To:     time.Now().AddDate(0, 0, c.lookaheadDays),
		HtmlId: pkg.NameToId(c.Name()),
	}, nil
}

func buildCard(orig vcard.Card, date *vcard.Field, anniversaryType string, now time.Time) (Card, error) {
	ret := Card{}

	if orig == nil || date == nil {
		return ret, errors.New("buildCard: nil parameter supplied")
	}

	for _, name := range orig.Names() {
		ret.Name = fmt.Sprintf("%s %s", name.GivenName, name.FamilyName)
		break
	}

	var err error
	ret.Anniversary, err = parseTimeCard(date.Value)
	year := now.Year()
	if ret.Anniversary.Month() < now.Month() {
		year += 1
	}
	ret.Upcoming = time.Date(year, ret.Anniversary.Month(), ret.Anniversary.Day(), 0, 0, 0, 0, time.UTC)
	if err != nil {
		return ret, err
	}

	ret.DateFormatted = getFormattedAnniversaryDate(ret.Anniversary, now)
	ret.Type = anniversaryType
	ret.Years = now.Year() - ret.Anniversary.Year()
	return ret, nil
}
