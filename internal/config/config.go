package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Email       *EmailConfig                `yaml:"email"`
	Http        *HttpConfig                 `yaml:"http"`
	Datasources []DatasourceConfigContainer `yaml:"datasources"`
}

func ReadConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	conf := DefaultConfig()
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func DefaultConfig() Config {
	c := Config{
		Http:        DefaultHttpConfig(),
		Datasources: nil,
	}

	return c
}

func DefaultHttpConfig() *HttpConfig {
	return &HttpConfig{
		Address:   "0.0.0.0:8080",
		ServePath: "/",
	}
}

type HttpConfig struct {
	Address   string `yaml:"address" validate:"omitempty,hostname_port"`
	ServePath string `yaml:"path" validate:"omitempty,filepath"`
}

func DefaultEmailConfig() EmailConfig {
	return EmailConfig{
		IsUtc: true,
	}
}

type EmailConfig struct {
	At          string `yaml:"at" validate:"datetime=03:04"`
	IsUtc       bool   `yaml:"is_utc"`
	SendAtStart bool   `yaml:"send_at_start"`

	UserName string   `yaml:"username"`
	Password string   `yaml:"password"`
	From     string   `yaml:"from" validate:"required"`
	To       []string `yaml:"to" validate:"required"`
	Host     string   `yaml:"host" validate:"hostname_port"`
}
