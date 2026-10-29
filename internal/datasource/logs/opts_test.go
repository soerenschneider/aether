package logs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/soerenschneider/aether/internal/templates"
)

type staticQuerier []LogEntry

func (s staticQuerier) QueryVictorialogs(_ context.Context, _ VictorialogsQuery) ([]LogEntry, error) {
	return s, nil
}

func TestWithTemplateFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "logs.html")
	// uses a function of the datasource, which must be available to custom templates
	if err := os.WriteFile(file, []byte(`custom:{{ range . }}{{ .Message }}{{ end }}{{ clock "invalid" }}`), 0o600); err != nil {
		t.Fatal(err)
	}

	tmpl := templates.TemplateData{DefaultTemplate: []byte("default"), SimpleTemplate: []byte("simple")}
	ds, err := New("http://localhost", tmpl, WithTemplateFile(file), WithQuerier(staticQuerier{{Message: "boom"}}))
	if err != nil {
		t.Fatal(err)
	}

	data, err := ds.GetData(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data.RenderedDefaultTemplate); got != "custom:boominvalid" {
		t.Errorf("expected the custom template to be used, got %q", got)
	}

	if _, err := New("http://localhost", tmpl, WithTemplateFile(filepath.Join(t.TempDir(), "missing.html"))); err == nil {
		t.Error("expected an error for a missing template file")
	}
}
