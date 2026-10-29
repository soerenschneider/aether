package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/soerenschneider/aether/pkg"
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

	return pkg.ReadSecretFile(c.UsernameFile)
}

func (c *EmailConfig) GetPassword() (string, error) {
	if len(c.Password) > 0 {
		return c.Password, nil
	}

	return pkg.ReadSecretFile(c.PasswordFile)
}

func (c *EmailConfig) GetFrom() (string, error) {
	if len(c.From) > 0 {
		return c.From, nil
	}

	return pkg.ReadSecretFile(c.FromFile)
}

func (c *EmailConfig) GetRecipients() ([]string, error) {
	if len(c.Recipients) > 0 {
		return c.Recipients, nil
	}

	data, err := os.ReadFile(c.RecipientsFile)
	if err != nil {
		return nil, err
	}

	// separated by commas or line breaks
	var recipients []string
	for _, recipient := range strings.FieldsFunc(string(data), func(r rune) bool { return r == ',' || r == '\n' || r == '\r' }) {
		if recipient = strings.TrimSpace(recipient); recipient != "" {
			recipients = append(recipients, recipient)
		}
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("no recipients in %q", c.RecipientsFile)
	}
	return recipients, nil
}
