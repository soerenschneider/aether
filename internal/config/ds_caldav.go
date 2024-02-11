package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type CalDavConfig struct {
	Endpoint string `yaml:"endpoint"`

	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`

	TemplateFile string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached       bool          `yaml:"cached"`
	CacheExpiry  time.Duration `yaml:"cache_expiry"`
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
