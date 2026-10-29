package templates

import (
	"embed"
	"errors"
	"io/fs"
	"os"
)

const migrationsDir = "migrations"

var (
	//go:embed */*.html */*.css
	migrations embed.FS

	// source the templates are read from, the embedded templates unless UseDir is called
	source fs.FS = migrations
)

// UseDir reads templates from a directory on disk instead of the embedded copies, so they can be edited without
// rebuilding. Only meant for development, must be called before any template is read.
func UseDir(dir string) {
	source = os.DirFS(dir)
}

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
	return fs.ReadFile(source, file)
}
