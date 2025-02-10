package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type StocksConfig struct {
	Symbols []string `yaml:"symbols"`

	TemplateFile       string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached             bool          `yaml:"cached"`
	CacheExpiry        time.Duration `yaml:"cache_expiry"`
	ExcludeFromSummary bool          `yaml:"exclude_from_summary"`
}

func (ds *StocksConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp StocksConfig

	conf := &tmp{
		Cached:      true,
		CacheExpiry: 15 * time.Minute,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = StocksConfig(*conf)
	return nil
}

func (ds *StocksConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}

func (ds *StocksConfig) IsCached() bool {
	return ds.Cached
}

func (ds *StocksConfig) Type() string {
	return Stocks
}
