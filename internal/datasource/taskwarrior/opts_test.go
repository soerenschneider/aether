package taskwarrior

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/soerenschneider/aether/internal/templates"
)

type staticClient []Task

func (s staticClient) GetTasks(_ context.Context) ([]Task, error) {
	return s, nil
}

func TestWithTemplateFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tasks.html")
	// uses a function of the datasource, which must be available to custom templates
	if err := os.WriteFile(file, []byte(`custom:{{ range .Tasks }}{{ .Description }} {{ formatDue .Due }}{{ end }}`), 0o600); err != nil {
		t.Fatal(err)
	}

	client := staticClient{{Description: "review", Due: time.Now().Add(time.Hour), Status: "pending"}}
	ds, err := New(client, templates.TemplateData{DefaultTemplate: []byte("default")}, WithTemplateFile(file))
	if err != nil {
		t.Fatal(err)
	}

	data, err := ds.GetData(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data.RenderedDefaultTemplate); !strings.HasPrefix(got, "custom:review ") {
		t.Errorf("expected the custom template to be used, got %q", got)
	}

	if _, err := New(client, templates.TemplateData{DefaultTemplate: []byte("default")}, WithTemplateFile(filepath.Join(t.TempDir(), "missing.html"))); err == nil {
		t.Error("expected an error for a missing template file")
	}
}
