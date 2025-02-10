package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/multierr"
)

const defaultUnit = "metric"
const defaultOpenWeatherApiUrl = "https://api.openweathermap.org/data/2.5"
const defaultCount = 10

type OpenweatherMapClient struct {
	httpClient *http.Client
	apiKey     string
	baseUrl    string
	units      string
	lat        Lat
	lon        Lon
	niceName   string
	count      int
}

type OpenweatherMapOpt func(client *OpenweatherMapClient) error

func NewOpenweatherMapClient(apiKey string, lat Lat, lon Lon, niceName string, opts ...OpenweatherMapOpt) (*OpenweatherMapClient, error) {
	ds := &OpenweatherMapClient{
		apiKey:     apiKey,
		lat:        lat,
		lon:        lon,
		niceName:   niceName,
		baseUrl:    defaultOpenWeatherApiUrl,
		count:      defaultCount,
		units:      defaultUnit,
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

func (w *OpenweatherMapClient) GetLocation() (Lat, Lon) {
	return w.lat, w.lon
}

func (w *OpenweatherMapClient) GetNiceName() string {
	return w.niceName
}

func (w *OpenweatherMapClient) GetWeatherData(ctx context.Context) (*WeatherData, error) {
	url := fmt.Sprintf("%s/forecast?lat=%f&lon=%f&cnt=%d&units=%s&appid=%s", w.baseUrl, w.lat, w.lon, w.count, w.units, w.apiKey)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := w.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

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
