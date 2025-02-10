package weather

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"github.com/soerenschneider/aether/pkg"
)

type Lat float64
type Lon float64

type Client interface {
	GetWeatherData(ctx context.Context) (*WeatherData, error)
	// Lat, lon
	GetLocation() (Lat, Lon)
	GetNiceName() string
}

type WeatherDatasource struct {
	client             Client
	regularTemplate    *template.Template
	simpleTemplate     *template.Template
	excludeFromSummary bool
}

func New(client Client, templateData templates.TemplateData) (*WeatherDatasource, error) {
	if client == nil {
		return nil, errors.New("nil client passed")
	}
	ds := &WeatherDatasource{
		client: client,
	}

	funcMap := template.FuncMap{
		"weekday":               formatWeekday,
		"getClassForClouds":     getClassForClouds,
		"getClassForHumidity":   getClassForHumidity,
		"getClassForPop":        getClassForPop,
		"getClassForRain":       getClassForRain,
		"getClassForTemp":       getClassForTemp,
		"getClassForVisibility": getClassForVisibility,
		"getClassForWind":       getClassForWind,
	}

	var err error
	if templateData.DefaultTemplate != nil {
		ds.regularTemplate, err = template.New("weather-regular").Funcs(funcMap).Parse(string(templateData.DefaultTemplate))
		if err != nil {
			return nil, err
		}
	}

	if templateData.SimpleTemplate != nil {
		ds.simpleTemplate, err = template.New("weather-simple").Funcs(funcMap).Parse(string(templateData.SimpleTemplate))
		if err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (w *WeatherDatasource) Name() string {
	if w.client.GetNiceName() != "" {
		return fmt.Sprintf("Weather %s", w.client.GetNiceName())
	}
	lat, lon := w.client.GetLocation()
	return fmt.Sprintf("Weather lat %f, lon %f", lat, lon)
}

func formatWeekday(t time.Time) string {
	return t.Weekday().String()
}

func (w *WeatherDatasource) GetData(ctx context.Context) (*internal.Data, error) {
	data, err := w.client.GetWeatherData(ctx)
	if err != nil {
		return nil, err
	}

	// set times for template
	now := time.Now()
	data.Now = now.Format("2006-01-02")
	data.Tomorrow = now.AddDate(0, 0, 1).Format("2006-01-02")
	data.HtmlId = pkg.NameToId(w.Name())

	var regularTemplateData bytes.Buffer
	if err := w.regularTemplate.Execute(&regularTemplateData, data); err != nil {
		return nil, err
	}

	var simpleTemplateData bytes.Buffer
	if w.simpleTemplate != nil {
		if err := w.simpleTemplate.Execute(&simpleTemplateData, data); err != nil {
			return nil, err
		}
	}

	var summary []string
	if !w.excludeFromSummary {
		summary = GenerateWeatherReport(data.List, time.Now())
	}

	return &internal.Data{
		Summary:                    summary,
		RenderedDefaultTemplate:    regularTemplateData.Bytes(),
		RenderedSimplifiedTemplate: simpleTemplateData.Bytes(),
	}, nil
}
