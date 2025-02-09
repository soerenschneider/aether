package alertmanager

import (
	"bytes"
	"context"
	"html/template"
	"sort"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"go.uber.org/multierr"
	"golang.org/x/exp/maps"
)

const defaultLimit = 10

type Alert struct {
	Name     string
	Severity string
	Count    int
}

type AlertmanagerDatasource struct {
	client   alert.ClientService
	basePath string
	scheme   string

	defaultTemplate *template.Template
	simpleTemplate  *template.Template
	limit           int
}

type Opt func(datasource *AlertmanagerDatasource) error

func New(host string, templateData templates.TemplateData, opts ...Opt) (*AlertmanagerDatasource, error) {
	strf := strfmt.NewFormats()

	if err := templateData.Validate(); err != nil {
		return nil, err
	}

	ds := &AlertmanagerDatasource{
		limit:    defaultLimit,
		scheme:   "http",
		basePath: "/api/v2",
	}

	var errs error
	for _, opt := range opts {
		if err := opt(ds); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	transportConf := client.DefaultTransportConfig().
		WithHost(host).
		WithSchemes([]string{ds.scheme}).
		WithBasePath(ds.basePath)

	apiClient := client.NewHTTPClientWithConfig(
		strf,
		transportConf,
	)

	alertClient := alert.New(apiClient.Transport, strf)
	ds.client = alertClient

	var err error
	ds.defaultTemplate, err = template.New("alertmanager-default").Parse(string(templateData.DefaultTemplate))
	if err != nil {
		return nil, err
	}

	if len(templateData.SimpleTemplate) > 0 {
		ds.simpleTemplate, err = template.New("alertmanager-simple").Parse(string(templateData.SimpleTemplate))
		if err != nil {
			return nil, err
		}
	}

	return ds, errs
}

func (a *AlertmanagerDatasource) Name() string {
	return "Alertmanager"
}

func (a *AlertmanagerDatasource) GetData(ctx context.Context) (*internal.Data, error) {
	alerts, err := a.getActiveAlerts(ctx)
	if err != nil {
		return nil, err
	}

	sortAlerts(alerts)
	if len(alerts) > a.limit {
		alerts = alerts[0:a.limit]
	}

	var renderedDefaultTemplate bytes.Buffer
	if err := a.defaultTemplate.Execute(&renderedDefaultTemplate, alerts); err != nil {
		return nil, err
	}

	var renderedSimpleTemplate bytes.Buffer
	if a.simpleTemplate != nil {
		if err := a.defaultTemplate.Execute(&renderedSimpleTemplate, alerts); err != nil {
			return nil, err
		}
	}

	ret := &internal.Data{
		Summary:                    nil,
		RenderedDefaultTemplate:    renderedDefaultTemplate.Bytes(),
		RenderedSimplifiedTemplate: renderedSimpleTemplate.Bytes(),
	}
	return ret, nil
}

var severities = map[string]int{
	"critical": 3,
	"major":    2,
	"warning":  1,
}

func sortAlerts(alerts []*Alert) {
	sort.Slice(alerts, func(i, j int) bool {
		a := severities[alerts[i].Severity]
		b := severities[alerts[j].Severity]
		if a == b {
			return alerts[i].Count > alerts[j].Count
		}

		return a > b
	})
}

func (a *AlertmanagerDatasource) getActiveAlerts(ctx context.Context) ([]*Alert, error) {
	resp, err := a.client.GetAlerts(alert.NewGetAlertsParamsWithContext(ctx))
	if err != nil {
		return nil, err
	}

	if resp.GetPayload() == nil {
		return nil, nil
	}

	alerts := map[string]*Alert{}
	for _, alert := range resp.GetPayload() {
		if alert.Status != nil && alert.Status.State != nil && *alert.Status.State == "active" {
			name := alert.Labels["alertname"]

			a, ok := alerts[name]
			if !ok {
				alerts[name] = &Alert{
					Name:     name,
					Severity: alert.Labels["severity"],
					Count:    1,
				}
			} else {
				a.Count += 1
			}
		}
	}

	return maps.Values(alerts), nil
}
