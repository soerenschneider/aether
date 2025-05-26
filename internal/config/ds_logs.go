package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type LogsConfig struct {
	Endpoint string `yaml:"endpoint" validate:"required,url"`
	Query    string `yaml:"query"`
	Limit    int    `yaml:"limit" validate:"omitempty,gte=1,lte=50"`

	TemplateFile       string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached             bool          `yaml:"cached"`
	CacheExpiry        time.Duration `yaml:"cache_expiry"`
	ExcludeFromSummary bool          `yaml:"exclude_from_summary"`
}

func (ds *LogsConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp LogsConfig

	conf := &tmp{
		Cached:      true,
		CacheExpiry: 1 * time.Minute,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = LogsConfig(*conf)
	return nil
}

func (ds *LogsConfig) Type() string {
	return Logs
}

func (ds *LogsConfig) IsCached() bool {
	return ds.Cached
}

func (ds *LogsConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}
