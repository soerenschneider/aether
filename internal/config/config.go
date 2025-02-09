package config

import (
	"os"
	"strings"

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
		UseGzip:   true,
		Minify:    true,
	}
}

type HttpConfig struct {
	Address              string `yaml:"address" validate:"omitempty,hostname_port"`
	ServePath            string `yaml:"path" validate:"omitempty,filepath"`
	Minify               bool   `yaml:"minify"`
	UseGzip              bool   `yaml:"use_gzip"`
	GzipCompressionLevel int    `yaml:"gzip" validate:"gt=-2,lt=10"`
}

type EmailConfig struct {
	At       string `yaml:"at" validate:"datetime=03:04"`
	Timezone string `yaml:"timezone" validate:"timezone"`

	Username     string `yaml:"username"`
	UsernameFile string `yaml:"username_file"`
	Password     string `yaml:"password"`
	PasswordFile string `yaml:"password_file"`

	From           string   `yaml:"from" validate:"required_without=FromFile"`
	FromFile       string   `yaml:"from_file" validate:"required_without=From"`
	Recipients     []string `yaml:"recipient" validate:"required_without=RecipientsFile"`
	RecipientsFile string   `yaml:"recipient_file" validate:"required_without=Recipients"`
	Host           string   `yaml:"host" validate:"hostname_port"`
}

func (c *EmailConfig) GetUsername() (string, error) {
	if len(c.Username) > 0 {
		return c.Username, nil
	}

	data, err := os.ReadFile(c.UsernameFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *EmailConfig) GetPassword() (string, error) {
	if len(c.Password) > 0 {
		return c.Password, nil
	}

	data, err := os.ReadFile(c.PasswordFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *EmailConfig) GetFrom() (string, error) {
	if len(c.From) > 0 {
		return c.From, nil
	}

	data, err := os.ReadFile(c.FromFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *EmailConfig) GetRecipients() ([]string, error) {
	if len(c.Recipients) > 0 {
		return c.Recipients, nil
	}

	data, err := os.ReadFile(c.RecipientsFile)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), ","), nil
}
