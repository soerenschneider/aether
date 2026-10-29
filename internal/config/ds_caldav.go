package config

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type CalDavConfig struct {
	// the account, used by all calendars that don't specify their own
	CalDavAccount `yaml:",inline"`

	// calendars to show merged in a single agenda. If empty, the first calendar of the account is used.
	Calendars []CalDavCalendarConfig `yaml:"calendars" validate:"dive"`

	TemplateFile       string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached             bool          `yaml:"cached"`
	CacheExpiry        time.Duration `yaml:"cache_expiry"`
	ExcludeFromSummary bool          `yaml:"exclude_from_summary"`
}

type CalDavAccount struct {
	Endpoint string `yaml:"endpoint"`

	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`
}

type CalDavCalendarConfig struct {
	// shown next to the entries of this calendar
	Name string `yaml:"name" validate:"required"`
	// name or path of the calendar within the account, the first calendar of the account if empty
	Calendar string `yaml:"calendar"`

	// optional, the account of the datasource is used if no endpoint is given
	CalDavAccount `yaml:",inline"`
}

// GetCalendars returns the calendars to fetch, each with its account. Calendars without an endpoint use the account
// of the datasource.
func (ds *CalDavConfig) GetCalendars() ([]CalDavCalendarConfig, error) {
	if len(ds.Calendars) == 0 {
		if ds.Endpoint == "" {
			return nil, errors.New("caldav: no endpoint configured")
		}
		return []CalDavCalendarConfig{{CalDavAccount: ds.CalDavAccount}}, nil
	}

	seen := map[string]bool{}
	ret := make([]CalDavCalendarConfig, 0, len(ds.Calendars))
	for _, calendar := range ds.Calendars {
		if seen[calendar.Name] {
			return nil, fmt.Errorf("caldav: calendar name %q is used more than once", calendar.Name)
		}
		seen[calendar.Name] = true

		if calendar.Endpoint == "" {
			if ds.Endpoint == "" {
				return nil, fmt.Errorf("caldav: calendar %q has no endpoint and there's no default endpoint configured", calendar.Name)
			}
			calendar.CalDavAccount = ds.CalDavAccount
		}
		ret = append(ret, calendar)
	}

	return ret, nil
}

func (ds *CalDavConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp CalDavConfig

	conf := &tmp{
		Cached:      true,
		CacheExpiry: 30 * time.Minute,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = CalDavConfig(*conf)
	return nil
}

func (ds *CalDavConfig) Type() string {
	return CalDav
}

func (ds *CalDavConfig) IsCached() bool {
	return ds.Cached
}

func (ds *CalDavConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}
