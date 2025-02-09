package templates

import (
	"embed"
	"errors"
	"io/fs"
)

const migrationsDir = "migrations"

var (
	//go:embed */*.html
	migrations embed.FS
)

type TemplateData struct {
	DefaultTemplate []byte
	SimpleTemplate  []byte
}

func (t *TemplateData) Validate() error {
	if len(t.DefaultTemplate) == 0 {
		return errors.New("no default template defined")
	}
	return nil
}

func GetTemplates() ([]fs.DirEntry, error) {
	return migrations.ReadDir(migrationsDir)
}

func GetTemplate(file string) ([]byte, error) {
	return migrations.ReadFile(file)
}
