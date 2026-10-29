package alertmanager

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/soerenschneider/aether/internal/templates"
)

type staticAlerts struct{}

func (staticAlerts) GetAlerts(_ *alert.GetAlertsParams, _ ...alert.ClientOption) (*alert.GetAlertsOK, error) {
	active := "active"
	return &alert.GetAlertsOK{Payload: models.GettableAlerts{{
		Alert:  models.Alert{Labels: models.LabelSet{"alertname": "DiskSpaceLow", "severity": "critical"}},
		Status: &models.AlertStatus{State: &active},
	}}}, nil
}

func (staticAlerts) PostAlerts(_ *alert.PostAlertsParams, _ ...alert.ClientOption) (*alert.PostAlertsOK, error) {
	return nil, errors.New("not supported")
}

func (staticAlerts) SetTransport(_ runtime.ClientTransport) {}

func TestWithTemplateFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "alerts.html")
	if err := os.WriteFile(file, []byte(`custom:{{ range . }}{{ .Name }}{{ end }}`), 0o600); err != nil {
		t.Fatal(err)
	}

	tmpl := templates.TemplateData{DefaultTemplate: []byte("default"), SimpleTemplate: []byte(`simple:{{ len . }}`)}
	ds, err := New("localhost", tmpl, WithTemplateFile(file), WithClient(staticAlerts{}))
	if err != nil {
		t.Fatal(err)
	}

	data, err := ds.GetData(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data.RenderedDefaultTemplate); got != "custom:DiskSpaceLow" {
		t.Errorf("expected the custom template to be used, got %q", got)
	}
	if got := string(data.RenderedSimplifiedTemplate); got != "simple:1" {
		t.Errorf("expected the simple template to be used, got %q", got)
	}

	if _, err := New("localhost", tmpl, WithTemplateFile(filepath.Join(t.TempDir(), "missing.html"))); err == nil {
		t.Error("expected an error for a missing template file")
	}
}
