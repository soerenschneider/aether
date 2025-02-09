package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type WeatherConfig struct {
	Latitude   float64 `yaml:"latitude" validate:"latitude"`
	Longitude  float64 `yaml:"longitude" validate:"longitude"`
	ApiKey     string  `yaml:"apikey" validate:"required_without=ApiKeyFile"`
	ApiKeyFile string  `yaml:"apikey_file" validate:"required_without=ApiKey"`

	TemplateFile string        `yaml:"template_file" validate:"omitempty,filepath"`
	Cached       bool          `yaml:"cached"`
	CacheExpiry  time.Duration `yaml:"cache_expiry"`
	NiceName     string        `yaml:"nice_name"`
	Count        int           `yaml:"count"`
}

func (ds *WeatherConfig) UnmarshalYAML(node *yaml.Node) error {
	type tmp WeatherConfig

	conf := &tmp{
		Cached:      true,
		CacheExpiry: 15 * time.Minute,
	}
	if err := node.Decode(&conf); err != nil {
		return err
	}

	*ds = WeatherConfig(*conf)
	return nil
}

func (ds *WeatherConfig) Type() string {
	return Weather
}

func (ds *WeatherConfig) IsCached() bool {
	return ds.Cached
}

func (ds *WeatherConfig) GetCacheExpiry() time.Duration {
	return ds.CacheExpiry
}
