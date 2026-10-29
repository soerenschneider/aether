package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCalDavConfig_GetCalendars(t *testing.T) {
	input := `
type: caldav
endpoint: https://cloud.example.org/remote.php/dav
username: soeren
password_file: /run/secrets/caldav
calendars:
  - name: Private
    calendar: personal
  - name: Family
    calendar: Family
  - name: Work
    endpoint: https://other.example.org/dav
    username: work
    password: secret
`
	var container DatasourceConfigContainer
	if err := yaml.Unmarshal([]byte(input), &container); err != nil {
		t.Fatal(err)
	}

	conf := container.Config.(*CalDavConfig)
	if !conf.Cached {
		t.Error("expected the default to be kept")
	}

	calendars, err := conf.GetCalendars()
	if err != nil {
		t.Fatal(err)
	}
	if len(calendars) != 3 {
		t.Fatalf("expected 3 calendars, got %d", len(calendars))
	}

	for _, calendar := range calendars[:2] {
		if calendar.Endpoint != "https://cloud.example.org/remote.php/dav" || calendar.Username != "soeren" || calendar.PasswordFile != "/run/secrets/caldav" {
			t.Errorf("calendar %q did not inherit the account: %+v", calendar.Name, calendar)
		}
	}
	if calendars[0].Calendar != "personal" || calendars[1].Calendar != "Family" {
		t.Errorf("unexpected calendars: %+v", calendars)
	}
	if work := calendars[2]; work.Endpoint != "https://other.example.org/dav" || work.Username != "work" || work.Password != "secret" || work.PasswordFile != "" {
		t.Errorf("calendar Work should use its own account: %+v", work)
	}
}

func TestCalDavConfig_GetCalendars_legacy(t *testing.T) {
	conf := &CalDavConfig{CalDavAccount: CalDavAccount{Endpoint: "https://cloud.example.org", Username: "soeren"}}
	calendars, err := conf.GetCalendars()
	if err != nil {
		t.Fatal(err)
	}
	if len(calendars) != 1 || calendars[0].Name != "" || calendars[0].Endpoint != "https://cloud.example.org" || calendars[0].Calendar != "" {
		t.Errorf("unexpected calendars: %+v", calendars)
	}
}

func TestCalDavConfig_GetCalendars_invalid(t *testing.T) {
	tests := map[string]*CalDavConfig{
		"no endpoint at all": {},
		"calendar without endpoint": {Calendars: []CalDavCalendarConfig{
			{Name: "Private"},
		}},
		"duplicate names": {CalDavAccount: CalDavAccount{Endpoint: "https://cloud.example.org"}, Calendars: []CalDavCalendarConfig{
			{Name: "Private"}, {Name: "Private", Calendar: "other"},
		}},
	}

	for name, conf := range tests {
		if _, err := conf.GetCalendars(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestCalDavConfig_calendarNameRequired(t *testing.T) {
	input := `
type: caldav
endpoint: https://cloud.example.org
calendars:
  - calendar: personal
`
	var container DatasourceConfigContainer
	if err := yaml.Unmarshal([]byte(input), &container); err == nil {
		t.Error("expected a validation error for a calendar without name")
	}
}
