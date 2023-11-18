package weather

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sync"

	"go.uber.org/multierr"
)

const defaultUnit = "metric"

type WeatherDatasource struct {
	units  string
	lat    string
	lon    string
	apiKey string

	httpClient *http.Client
	template   *template.Template
	once       sync.Once
}

type Opt func(datasource *WeatherDatasource) error

func New(apiKey string, lat string, lon string, opts ...Opt) (*WeatherDatasource, error) {
	ds := &WeatherDatasource{
		apiKey:     apiKey,
		units:      defaultUnit,
		lat:        lat,
		lon:        lon,
		httpClient: http.DefaultClient,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(ds); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return ds, errs
}

func (w *WeatherDatasource) Name() string {
	return fmt.Sprintf("Weather lat %s, lon %s", w.lat, w.lon)
}

func (w *WeatherDatasource) GetHtml(ctx context.Context) (string, error) {
	w.once.Do(func() {
		if w.template == nil {
			w.template = template.Must(template.New("weather").Parse(defaultTemplate))
		}
	})

	data, err := w.getWeatherData(ctx)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = w.template.Execute(&tpl, data)
	return tpl.String(), err
}

func (w *WeatherDatasource) getWeatherData(ctx context.Context) (*WeatherData, error) {
	url := fmt.Sprintf("%s?lat=%s&lon=%s&cnt=10&units=%s&appid=%s", apiUrl, w.lat, w.lon, w.units, w.apiKey)
	resp, err := w.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	weatherData := &WeatherData{}
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return nil, err
	}

	return weatherData, nil
}

const apiUrl = "https://api.openweathermap.org/data/2.5/forecast"
