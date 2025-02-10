package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type CardDavConfig struct {
	Endpoint string `yaml:"endpoint"`

	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`

	TemplateFile       string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached             bool          `yaml:"cached"`
	CacheExpiry        time.Duration `yaml:"cache_expiry"`
	ExcludeFromSummary bool          `yaml:"exclude_from_summary"`
}

func (ds *CardDavConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp CardDavConfig

	conf := &tmp{
		Cached:      true,
		CacheExpiry: 4 * time.Hour,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = CardDavConfig(*conf)
	return nil
}

func (ds *CardDavConfig) Type() string {
	return CardDav
}

func (ds *CardDavConfig) IsCached() bool {
	return ds.Cached
}

func (ds *CardDavConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}
