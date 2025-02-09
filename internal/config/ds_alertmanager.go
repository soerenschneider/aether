package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type AlertmanagerConfig struct {
	Host     string `yaml:"host"`
	BasePath string `yaml:"base_path"`
	Scheme   string `yaml:"scheme" validate:"oneof=http https"`

	TemplateFile       string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached             bool          `yaml:"cached"`
	CacheExpiry        time.Duration `yaml:"cache_expiry"`
	SimpleTemplateFile string        `yaml:"simple_template_file" validate:"omitempty,filepath"`
}

func (ds *AlertmanagerConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp AlertmanagerConfig

	conf := &tmp{
		Cached:      false,
		CacheExpiry: 1 * time.Minute,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = AlertmanagerConfig(*conf)
	return nil
}

func (ds *AlertmanagerConfig) Type() string {
	return Alertmanager
}

func (ds *AlertmanagerConfig) IsCached() bool {
	return ds.Cached
}

func (ds *AlertmanagerConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}
