//go:build dev

package main

// Fake data for the development server. The fakes replace the clients of the datasources, so filtering, sorting,
// summaries and css classes are produced by the real datasource code. All dates are relative to now.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/emersion/go-vcard"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/datasource/alertmanager"
	"github.com/soerenschneider/aether/internal/datasource/astral"
	"github.com/soerenschneider/aether/internal/datasource/caldav"
	"github.com/soerenschneider/aether/internal/datasource/carddav"
	"github.com/soerenschneider/aether/internal/datasource/logs"
	"github.com/soerenschneider/aether/internal/datasource/taskwarrior"
	"github.com/soerenschneider/aether/internal/datasource/weather"
	"github.com/soerenschneider/aether/internal/templates"
	"github.com/soerenschneider/aether/pkg"

	webdavcarddav "github.com/emersion/go-webdav/carddav"
	"github.com/go-openapi/runtime"
)

type scenario string

const (
	// a regular day
	scenarioNormal scenario = "normal"
	// lots of entries, long texts and extreme values that trigger every highlight
	scenarioBusy scenario = "busy"
	// nothing to show for any datasource that can be empty
	scenarioEmpty scenario = "empty"
	// some datasources can not be reached
	scenarioFailing scenario = "failing"
	// regular data, but the server goes away after the page has been loaded
	scenarioStale scenario = "stale"
	// lots happening today: several events and more than one anniversary
	scenarioPacked scenario = "packed"
)

var scenarios = map[scenario]struct{}{
	scenarioNormal:  {},
	scenarioBusy:    {},
	scenarioEmpty:   {},
	scenarioFailing: {},
	scenarioStale:   {},
	scenarioPacked:  {},
}

func scenarioNames() []string {
	ret := make([]string, 0, len(scenarios))
	for s := range scenarios {
		ret = append(ret, string(s))
	}
	sort.Strings(ret)
	return ret
}

var errSimulatedOutage = errors.New("simulated outage")

func buildFakeDatasources(s scenario, recorder *errorRecorder) ([]Datasource, error) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return nil, err
	}

	var datasources []Datasource
	add := func(ds Datasource, err error) error {
		if err != nil {
			return fmt.Errorf("building datasource: %w", err)
		}
		datasources = append(datasources, recorder.wrap(ds))
		return nil
	}

	builders := []func() error{
		func() error {
			tmpl, err := loadTemplates("weather/default.html", "weather/simple.html")
			if err != nil {
				return err
			}
			return add(weather.New(fakeWeather{scenario: s}, tmpl))
		},
		func() error {
			tmpl, err := loadTemplates("calendar/default.html", "calendar/simple.html")
			if err != nil {
				return err
			}
			// a single calendar, like a config without "calendars"
			calendars := []caldav.NamedCalendar{{Client: fakeCalendar{scenario: s}}}
			if s == scenarioPacked || s == scenarioFailing {
				calendars = []caldav.NamedCalendar{
					{Name: "Private", Client: fakeCalendar{scenario: s, calendar: calendarPrivate}},
					{Name: "Family", Client: fakeCalendar{scenario: s, calendar: calendarFamily}},
				}
			}
			client, err := caldav.NewMultiClient(calendars...)
			if err != nil {
				return err
			}
			return add(caldav.New(client, tmpl))
		},
		func() error {
			tmpl, err := loadTemplates("taskwarrior/default.html", "")
			if err != nil {
				return err
			}
			return add(taskwarrior.New(fakeTasks{scenario: s}, tmpl))
		},
		func() error {
			tmpl, err := loadTemplates("alertmanager/default.html", "")
			if err != nil {
				return err
			}
			return add(alertmanager.New("fake", tmpl, alertmanager.WithClient(fakeAlerts{scenario: s})))
		},
		func() error {
			tmpl, err := loadTemplates("astral/default.html", "")
			if err != nil {
				return err
			}
			return add(astral.New(52.52, 13.405, tmpl, berlin))
		},
		func() error {
			tmpl, err := loadTemplates("contacts/default.html", "contacts/simple.html")
			if err != nil {
				return err
			}
			return add(carddav.New("http://fake", tmpl, carddav.WithClient(fakeContacts{scenario: s})))
		},
		func() error {
			tmpl, err := loadTemplates("logs/default.html", "logs/simple.html")
			if err != nil {
				return err
			}
			return add(logs.New("http://fake", tmpl, logs.WithQuerier(fakeLogs{scenario: s})))
		},
	}

	for _, build := range builders {
		if err := build(); err != nil {
			return nil, err
		}
	}

	if s == scenarioFailing {
		// a second weather location that can't be reached, the family calendar fails as well (see fakeCalendar)
		datasources = append(datasources, failingDatasource{name: "Weather Hamburg"})
	}

	return datasources, nil
}

func loadTemplates(defaultTemplate, simpleTemplate string) (templates.TemplateData, error) {
	var err error
	ret := templates.TemplateData{}
	ret.DefaultTemplate, err = templates.GetTemplate(defaultTemplate)
	if err != nil {
		return ret, err
	}

	if simpleTemplate != "" {
		ret.SimpleTemplate, err = templates.GetTemplate(simpleTemplate)
	}
	return ret, err
}

type failingDatasource struct {
	name string
}

func (f failingDatasource) Name() string {
	return f.name
}

func (f failingDatasource) GetData(_ context.Context) (*internal.Data, error) {
	return nil, errSimulatedOutage
}

// today returns today's date at the given hour and minute.
func today(hour, minute int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
}

func inDays(days, hour, minute int) time.Time {
	return today(hour, minute).AddDate(0, 0, days)
}

// ---------------------------------------------------------------------------------------------------------------------
// weather

type fakeWeather struct {
	scenario scenario
}

func (f fakeWeather) GetLocation() (weather.Lat, weather.Lon) {
	return 52.52, 13.405
}

func (f fakeWeather) GetNiceName() string {
	return "Berlin"
}

type fakeCondition struct {
	emoji, description string
}

var (
	conditionClear        = fakeCondition{"☀️", "clear sky"}
	conditionFewClouds    = fakeCondition{"☁️", "few clouds"}
	conditionOvercast     = fakeCondition{"☁️", "overcast clouds"}
	conditionLightRain    = fakeCondition{"☔️", "light rain"}
	conditionHeavyRain    = fakeCondition{"☔️", "heavy intensity rain"}
	conditionThunderstorm = fakeCondition{"⚡️", "thunderstorm with heavy rain"}
	conditionSnow         = fakeCondition{"❄️", "light snow"}
	conditionFog          = fakeCondition{"🌪", "fog"}
)

func (f fakeWeather) GetWeatherData(_ context.Context) (*weather.WeatherData, error) {
	now := time.Now()
	// forecasts come in 3h slots
	first := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()/3*3, 0, 0, 0, now.Location()).Add(3 * time.Hour)

	data := &weather.WeatherData{
		City: weather.City{
			ID:          2950159,
			Name:        "Berlin",
			SunriseTime: today(7, 12),
			SunsetTime:  today(16, 48),
		},
	}

	for i := 0; i < 24; i++ {
		slot := first.Add(time.Duration(i*3) * time.Hour)
		var entry *weather.WeatherEntry
		if f.scenario == scenarioBusy {
			entry = extremeWeather(i)
		} else {
			entry = regularWeather(slot, i)
		}
		entry.Time = slot
		entry.Dt = int(slot.Unix())
		entry.Wind.Direction, entry.Wind.DirectionEmoji = pkg.TranslateDegreeToDirection(float64(entry.Wind.Deg))
		data.List = append(data.List, entry)
	}

	return data, nil
}

// regularWeather is a mild autumn day with some rain in the afternoon.
func regularWeather(slot time.Time, i int) *weather.WeatherEntry {
	// coldest at 05:00, warmest at 14:00
	temp := 9 + 5*math.Sin(float64(slot.Hour()-8)/24*2*math.Pi)
	conditions := []fakeCondition{conditionClear, conditionFewClouds, conditionOvercast, conditionLightRain, conditionOvercast, conditionFewClouds, conditionClear, conditionClear}
	pops := []float64{0, 0.05, 0.2, 0.45, 0.3, 0.1, 0, 0}
	rains := []float64{0, 0, 0, 0.8, 0.2, 0, 0, 0}
	clouds := []int{5, 30, 85, 95, 70, 35, 10, 0}

	entry := &weather.WeatherEntry{}
	slotOfDay := i % len(conditions)
	entry.WeatherEmoji = conditions[slotOfDay].emoji
	entry.WeatherDescription = conditions[slotOfDay].description
	entry.Main.Temp = temp
	entry.Main.FeelsLike = temp - 1.5
	entry.Main.TempMin = temp - 1
	entry.Main.TempMax = temp + 1
	entry.Main.Humidity = 60 + 4*slotOfDay
	entry.Pop = pops[slotOfDay]
	entry.Rain.H3 = rains[slotOfDay]
	entry.Wind.Speed = 2 + float64(slotOfDay%4)
	entry.Wind.Deg = 220 + 10*slotOfDay
	entry.Clouds.All = clouds[slotOfDay]
	entry.VisibilityPercent = 100
	return entry
}

// extremeWeather cycles through values that trigger each of the css classes of the weather table.
func extremeWeather(i int) *weather.WeatherEntry {
	conditions := []fakeCondition{conditionThunderstorm, conditionHeavyRain, conditionSnow, conditionFog, conditionClear, conditionLightRain}
	temps := []float64{36, 31, 26, 22, 15, 4, -6}
	pops := []float64{1, 0.8, 0.6, 0.4, 0.1, 0}
	rains := []float64{55, 21, 8, 3, 0.4, 0}
	humidities := []int{95, 80, 50, 20}
	winds := []float64{24, 16, 11, 7, 2, 0}
	clouds := []int{100, 80, 60, 30, 0}
	visibilities := []int{10, 30, 60, 100}

	entry := &weather.WeatherEntry{}
	condition := conditions[i%len(conditions)]
	entry.WeatherEmoji = condition.emoji
	entry.WeatherDescription = condition.description
	entry.Main.Temp = temps[i%len(temps)]
	entry.Main.FeelsLike = temps[i%len(temps)] - 3
	entry.Main.TempMin = temps[i%len(temps)] - 1
	entry.Main.TempMax = temps[i%len(temps)] + 1
	entry.Main.Humidity = humidities[i%len(humidities)]
	entry.Pop = pops[i%len(pops)]
	entry.Rain.H3 = rains[i%len(rains)]
	entry.Wind.Speed = winds[i%len(winds)]
	entry.Wind.Deg = (i * 45) % 360
	entry.Clouds.All = clouds[i%len(clouds)]
	entry.VisibilityPercent = visibilities[i%len(visibilities)]
	return entry
}

// ---------------------------------------------------------------------------------------------------------------------
// calendar

const (
	calendarPrivate = "private"
	calendarFamily  = "family"
)

// entries of the family calendar, all others belong to the private calendar. Entries listed here with true are in
// both calendars.
var familyEntries = map[string]bool{
	"Moving day":       false,
	"Dinner with Alex": true,
	"Team offsite":     false,
	"Pick up parcel":   false,
	"Birthday party":   false,
	"Yoga":             false,
	"Lunch with Sam":   true,
}

type fakeCalendar struct {
	scenario scenario
	// private, family or empty for a single calendar that contains all entries
	calendar string
}

func (f fakeCalendar) FetchData(ctx context.Context) ([]caldav.Entry, error) {
	if f.scenario == scenarioFailing && f.calendar == calendarFamily {
		return nil, errSimulatedOutage
	}

	all, err := f.allEntries(ctx)
	if err != nil || f.calendar == "" {
		return all, err
	}

	var entries []caldav.Entry
	for _, entry := range all {
		inBoth, isFamily := familyEntries[entry.Summary]
		if inBoth || (isFamily == (f.calendar == calendarFamily)) {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (f fakeCalendar) allEntries(_ context.Context) ([]caldav.Entry, error) {
	var entries []caldav.Entry
	switch f.scenario {
	case scenarioEmpty:
		return nil, nil
	case scenarioBusy:
		titles := []string{
			"Quarterly planning with the whole infrastructure team including the folks from the Munich office",
			"Dentist",
			"Pick up parcel",
			"Flight BER → LIS",
			"Climbing",
			"Call with the landlord about the heating that has been making noises since September",
			"Birthday party",
		}
		locations := []string{
			"Konferenzraum 4.12 (Haus B, 4. OG), Musterstraße 123, 10115 Berlin",
			"",
			"Packstation 117",
			"Flughafen Berlin Brandenburg, Terminal 1",
			"Boulderklub Kreuzberg",
			"",
			"Tempelhofer Feld",
		}
		// more entries than the datasource displays, so the limit kicks in
		for i := 0; i < 14; i++ {
			start := inDays(i*2, 8+i%10, 0)
			entries = append(entries, caldav.NewEntry(titles[i%len(titles)], locations[i%len(locations)], start, start.Add(90*time.Minute)))
		}
	case scenarioPacked:
		now := time.Now()
		entries = []caldav.Entry{
			caldav.NewEntry("Team offsite", "", today(0, 0), inDays(1, 0, 0)),
			// already over
			caldav.NewEntry("Morning run", "", now.Add(-3*time.Hour), now.Add(-2*time.Hour)),
			// currently running
			caldav.NewEntry("Architecture review", "Room 4.12", now.Add(-30*time.Minute), now.Add(30*time.Minute)),
			caldav.NewEntry("Lunch with Sam", "Markthalle Neun, Berlin", now.Add(2*time.Hour), now.Add(3*time.Hour)),
			caldav.NewEntry("Dentist", "", now.Add(4*time.Hour), now.Add(5*time.Hour)),
			caldav.NewEntry("Yoga", "Studio Kreuzberg", now.Add(7*time.Hour), now.Add(8*time.Hour)),
			caldav.NewEntry("Dinner with Alex", "", inDays(1, 19, 0), inDays(1, 21, 30)),
		}
	default:
		entries = []caldav.Entry{
			caldav.NewEntry("Team standup", "", today(10, 0), today(10, 30)),
			caldav.NewEntry("Dinner with Alex", "Markthalle Neun, Berlin", inDays(1, 19, 0), inDays(1, 21, 30)),
			// whole day event
			caldav.NewEntry("Moving day", "", inDays(4, 0, 0), inDays(5, 0, 0)),
			caldav.NewEntry("Climbing", "Boulderklub Kreuzberg", inDays(8, 18, 0), inDays(8, 20, 0)),
			caldav.NewEntry("Dentist", "", inDays(16, 8, 30), inDays(16, 9, 15)),
		}
	}

	return entries, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// taskwarrior

type fakeTasks struct {
	scenario scenario
}

func (f fakeTasks) GetTasks(_ context.Context) ([]taskwarrior.Task, error) {
	now := time.Now()
	switch f.scenario {
	case scenarioEmpty:
		return nil, nil
	case scenarioPacked:
		return []taskwarrior.Task{
			// highest urgency, but not due soon: must not push the overdue tasks out of view
			{Id: 1, Description: "Plan the migration to the new NAS", Due: inDays(9, 0, 0), Project: "homelab", Urgency: 25, Status: "pending"},
			{Id: 2, Description: "Submit tax return", Due: inDays(-12, 0, 0), Project: "home.finance", Urgency: 12, Status: "pending", Tags: []string{"admin"}},
			{Id: 3, Description: "Reply to the landlord", Due: inDays(-1, 0, 0), Project: "home", Urgency: 8, Status: "pending"},
			{Id: 4, Description: "Renew TLS certificates on the edge proxy", Due: inDays(-4, 18, 0), Project: "homelab.infra", Urgency: 14, Status: "pending", Tags: []string{"ops"}},
			{Id: 5, Description: "Call the insurance", Due: today(9, 0), Project: "home", Urgency: 7, Status: "pending"},
			{Id: 6, Description: "Pick up the dry cleaning", Due: today(18, 0), Urgency: 4, Status: "pending"},
			{Id: 7, Description: "Review aether UI pull request", Due: today(0, 0), Project: "aether", Urgency: 9, Status: "pending", Tags: []string{"review"}},
			{Id: 8, Description: "Pay electricity bill", Due: inDays(1, 0, 0), Project: "home", Urgency: 6, Status: "pending"},
			{Id: 9, Description: "Prepare slides for Friday", Due: inDays(1, 10, 0), Project: "work", Urgency: 6.5, Status: "pending"},
			{Id: 10, Description: "Read chapter 4", Due: inDays(4, 0, 0), Urgency: 2, Status: "pending"},
			{Id: 11, Description: "Book dentist appointment", Urgency: 2.3, Status: "pending"},
			{Id: 12, Description: "Sort out old photos", Urgency: 1, Status: "pending"},
		}, nil
	case scenarioBusy:
		var tasks []taskwarrior.Task
		descriptions := []string{
			"Renew TLS certificates on the edge proxy before they expire and update the monitoring for the new expiry date",
			"Pay electricity bill",
			"Migrate the backups from restic to the new object storage bucket and verify that a restore actually works",
			"Read chapter 4",
			"Fix flaky integration test",
		}
		for i := 0; i < 18; i++ {
			task := taskwarrior.Task{
				Id:          int32(i + 1),
				Description: descriptions[i%len(descriptions)],
				Project:     []string{"homelab.infra.network", "home", "aether", ""}[i%4],
				Tags:        [][]string{{"ops", "urgent", "weekend", "needs-hardware"}, nil, {"review"}}[i%3],
				Urgency:     float32(20 - i),
				Status:      "pending",
			}
			// overdue, due today, in a few days, in a few weeks, no due date
			switch i % 5 {
			case 0:
				task.Due = now.Add(-30 * time.Hour)
			case 1:
				task.Due = now.Add(2 * time.Hour)
			case 2:
				task.Due = now.Add(60 * time.Hour)
			case 3:
				task.Due = now.Add(20 * 24 * time.Hour)
			}
			tasks = append(tasks, task)
		}
		return tasks, nil
	default:
		return []taskwarrior.Task{
			{Id: 3, Description: "Renew TLS certificates on the edge proxy", Due: now.Add(10 * time.Hour), Project: "homelab.infra", Urgency: 14.2, Status: "pending", Tags: []string{"ops"}},
			{Id: 7, Description: "Pay electricity bill", Due: now.Add(50 * time.Hour), Project: "home", Urgency: 9.8, Status: "pending"},
			{Id: 12, Description: "Review aether UI pull request", Due: now.Add(6 * 24 * time.Hour), Project: "aether", Urgency: 7.1, Status: "pending", Tags: []string{"review"}},
			{Id: 21, Description: "Book dentist appointment", Urgency: 2.3, Status: "pending"},
			// not shown: waiting and completed tasks are filtered out by the datasource
			{Id: 22, Description: "Waiting task", Urgency: 5, Status: "pending", Wait: now.Add(48 * time.Hour)},
			{Id: 23, Description: "Completed task", Urgency: 5, Status: "completed"},
		}, nil
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// alertmanager

type fakeAlerts struct {
	scenario scenario
}

func (f fakeAlerts) GetAlerts(_ *alert.GetAlertsParams, _ ...alert.ClientOption) (*alert.GetAlertsOK, error) {
	type fakeAlert struct {
		name, severity string
		count          int
	}

	var alerts []fakeAlert
	switch f.scenario {
	case scenarioEmpty:
	case scenarioBusy:
		// more distinct alerts than the datasource displays
		alerts = []fakeAlert{
			{"DiskSpaceLow", "critical", 3}, {"NodeDown", "critical", 1}, {"BackupStale", "warning", 2},
			{"CertificateExpiresSoon", "warning", 4}, {"HighMemoryUsage", "warning", 1}, {"ZFSPoolDegraded", "critical", 1},
			{"SmartctlDeviceTemperatureHighForAnUnusuallyLongAmountOfTime", "warning", 1}, {"UpsOnBattery", "critical", 1},
			{"PrometheusRuleFailures", "warning", 1}, {"TooManyRestarts", "info", 6}, {"NewVersionAvailable", "info", 2},
			{"Watchdog", "none", 1},
		}
	default:
		alerts = []fakeAlert{{"BackupStale", "warning", 1}, {"CertificateExpiresSoon", "info", 2}}
	}

	active := "active"
	suppressed := "suppressed"
	var payload models.GettableAlerts
	for _, a := range alerts {
		for i := 0; i < a.count; i++ {
			payload = append(payload, &models.GettableAlert{
				Alert:  models.Alert{Labels: models.LabelSet{"alertname": a.name, "severity": a.severity}},
				Status: &models.AlertStatus{State: &active},
			})
		}
	}
	// silenced alerts are ignored by the datasource
	payload = append(payload, &models.GettableAlert{
		Alert:  models.Alert{Labels: models.LabelSet{"alertname": "SilencedAlert", "severity": "critical"}},
		Status: &models.AlertStatus{State: &suppressed},
	})

	return &alert.GetAlertsOK{Payload: payload}, nil
}

func (f fakeAlerts) PostAlerts(_ *alert.PostAlertsParams, _ ...alert.ClientOption) (*alert.PostAlertsOK, error) {
	return nil, errors.New("not supported by fake")
}

func (f fakeAlerts) SetTransport(_ runtime.ClientTransport) {}

// ---------------------------------------------------------------------------------------------------------------------
// carddav

type fakeContacts struct {
	scenario scenario
}

func (f fakeContacts) FindAddressBookHomeSet(_ context.Context, _ string) (string, error) {
	return "/fake/", nil
}

func (f fakeContacts) FindAddressBooks(_ context.Context, _ string) ([]webdavcarddav.AddressBook, error) {
	return []webdavcarddav.AddressBook{{Path: "/fake/contacts/"}}, nil
}

func (f fakeContacts) QueryAddressBook(_ context.Context, _ string, _ *webdavcarddav.AddressBookQuery) ([]webdavcarddav.AddressObject, error) {
	type fakeContact struct {
		given, family string
		field         string
		// relative to today
		inDays, yearsAgo int
	}

	var contacts []fakeContact
	switch f.scenario {
	case scenarioEmpty:
		// out of the lookahead window, filtered by the datasource
		contacts = []fakeContact{{"Paul", "Weber", vcard.FieldBirthday, 40, 33}}
	case scenarioPacked:
		contacts = []fakeContact{
			{"Maria", "Schmidt", vcard.FieldBirthday, 0, 36},
			{"Jonas", "Müller", vcard.FieldAnniversary, 0, 8},
			{"Lea", "Fischer", vcard.FieldBirthday, 1, 29},
		}
	case scenarioBusy:
		given := []string{"Maria", "Jonas", "Lea", "Maximilian-Alexander", "Sophie", "Noah", "Emilia", "Ben"}
		family := []string{"Schmidt", "Müller", "Fischer", "von Hohenzollern-Sigmaringen", "Wagner", "Becker", "Hoffmann", "Schulz"}
		for i := 0; i < 10; i++ {
			field := vcard.FieldBirthday
			if i%3 == 0 {
				field = vcard.FieldAnniversary
			}
			contacts = append(contacts, fakeContact{given[i%len(given)], family[i%len(family)], field, i, 5 + i*7})
		}
	default:
		contacts = []fakeContact{
			{"Maria", "Schmidt", vcard.FieldBirthday, 1, 36},
			{"Jonas", "Müller", vcard.FieldAnniversary, 5, 8},
			{"Paul", "Weber", vcard.FieldBirthday, 40, 33},
		}
	}

	var ret []webdavcarddav.AddressObject
	for _, c := range contacts {
		card := vcard.Card{}
		card.SetName(&vcard.Name{GivenName: c.given, FamilyName: c.family})
		card.SetValue(c.field, inDays(c.inDays, 0, 0).AddDate(-c.yearsAgo, 0, 0).Format("2006-01-02"))
		ret = append(ret, webdavcarddav.AddressObject{Card: card})
	}
	return ret, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// logs

type fakeLogs struct {
	scenario scenario
}

func (f fakeLogs) QueryVictorialogs(_ context.Context, _ logs.VictorialogsQuery) ([]logs.LogEntry, error) {
	messages := []string{
		`level=error msg="could not reach upstream" host=nas.local err="dial tcp 10.0.0.12:443: i/o timeout"`,
		`level=error msg="backup job failed" job=restic-daily exit=1`,
		`level=error msg="certificate verification failed" err="x509: certificate has expired or is not yet valid"`,
	}

	count := 3
	switch f.scenario {
	case scenarioEmpty:
		return nil, nil
	case scenarioBusy:
		count = 20
		messages = append(messages,
			// long unbroken token to check that messages wrap
			`level=error msg="request failed" url=https://example.org/api/v1/very/long/path/that/does/not/contain/any/spaces/at/all?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0`,
			`panic: runtime error: invalid memory address or nil pointer dereference [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x4a2f1c] goroutine 1 [running]: main.main() /src/main.go:42 +0x1c`,
		)
	}

	var ret []logs.LogEntry
	for i := 0; i < count; i++ {
		ret = append(ret, logs.LogEntry{
			Timestamp: time.Now().Add(-time.Duration(i*7) * time.Minute).UTC().Format(time.RFC3339Nano),
			Message:   messages[i%len(messages)],
		})
	}
	return ret, nil
}
