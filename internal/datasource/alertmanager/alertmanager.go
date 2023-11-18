package alertmanager

import (
	"bytes"
	"context"
	"html/template"
	"sort"
	"sync"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
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

	once     sync.Once
	template *template.Template
	limit    int
}

type Opt func(datasource *AlertmanagerDatasource) error

func New(host string, opts ...Opt) (*AlertmanagerDatasource, error) {
	strf := strfmt.NewFormats()

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

	return ds, errs
}

func (a *AlertmanagerDatasource) Name() string {
	return "Alertmanager"
}

func (a *AlertmanagerDatasource) GetHtml(ctx context.Context) (string, error) {
	a.once.Do(func() {
		if a.template == nil {
			a.template = template.Must(template.New("alertmanager").Parse(defaultTemplate))
		}
	})

	alerts, err := a.getActiveAlerts(ctx)
	if err != nil {
		return "", err
	}

	sortAlerts(alerts)

	var tpl bytes.Buffer
	err = a.template.Execute(&tpl, alerts[:a.limit])
	return tpl.String(), err
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

func (d *AlertmanagerDatasource) getActiveAlerts(ctx context.Context) ([]*Alert, error) {
	resp, err := d.client.GetAlerts(alert.NewGetAlertsParamsWithContext(ctx))
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
