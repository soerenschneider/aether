package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type AstralConfig struct {
	Latitude  float64 `yaml:"latitude" validate:"latitude"`
	Longitude float64 `yaml:"longitude" validate:"longitude"`

	TemplateFile string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached       bool          `yaml:"cached"`
	CacheExpiry  time.Duration `yaml:"cache_expiry"`
}

func (ds *AstralConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp AstralConfig

	conf := &tmp{
		Cached:      true,
		CacheExpiry: 15 * time.Minute,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = AstralConfig(*conf)
	return nil
}

func (ds *AstralConfig) Type() string {
	return Astral
}

func (ds *AstralConfig) IsCached() bool {
	return ds.Cached
}

func (ds *AstralConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}
