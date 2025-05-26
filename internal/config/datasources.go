package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	Alertmanager = "alertmanager"
	Astral       = "astral"
	CalDav       = "caldav"
	CardDav      = "carddav"
	Logs         = "logs"
	Taskwarrior  = "taskwarrior"
	Stocks       = "stocks"
	Weather      = "weather"
)

type DatasourceConfigContainer struct {
	Config DatasourceConfig
}

func (ds *DatasourceConfigContainer) UnmarshalYAML(node *yaml.Node) error {
	type inner struct {
		Type string `yaml:"type"`
	}

	hookType := &inner{}
	if err := node.Decode(hookType); err != nil {
		return err
	}

	var conf DatasourceConfig
	switch hookType.Type {
	case Alertmanager:
		conf = &AlertmanagerConfig{}
	case Astral:
		conf = &AstralConfig{}
	case CalDav:
		conf = &CalDavConfig{}
	case CardDav:
		conf = &CardDavConfig{}
	case Logs:
		conf = &LogsConfig{}
	case Stocks:
		conf = &StocksConfig{}
	case Taskwarrior:
		conf = &TaskwarriorConfig{}
	case Weather:
		conf = &WeatherConfig{}

	default:
		return fmt.Errorf("unknown hook type %q", hookType.Type)
	}

	if err := node.Decode(conf); err != nil {
		return err
	}
	ds.Config = conf

	if err := Validate(ds.Config); err != nil {
		return err
	}

	return nil
}

var validate *validator.Validate = validator.New()

func Validate(s any) error {
	return validate.Struct(s)
}

type DatasourceConfig interface {
	Type() string
	IsCached() bool
	GetCacheExpiry() time.Duration
}
