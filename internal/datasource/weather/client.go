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
	//return getDummyWeather(w.niceName)
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

	//body := `{"cod":"200","message":0,"cnt":12,"list":[{"dt":1738702800,"main":{"temp":11.21,"feels_like":10.19,"temp_min":10.98,"temp_max":11.21,"pressure":1031,"sea_level":1031,"grnd_level":1024,"humidity":69,"temp_kf":0.23},"weather":[{"id":801,"main":"Clouds","description":"few clouds","icon":"02n"}],"clouds":{"all":12},"wind":{"speed":1.3,"deg":48,"gust":1.22},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-04 21:00:00"},{"dt":1738713600,"main":{"temp":10.36,"feels_like":9.25,"temp_min":9.88,"temp_max":10.36,"pressure":1032,"sea_level":1032,"grnd_level":1024,"humidity":69,"temp_kf":0.48},"weather":[{"id":802,"main":"Clouds","description":"scattered clouds","icon":"03n"}],"clouds":{"all":25},"wind":{"speed":1.8,"deg":70,"gust":1.72},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-05 00:00:00"},{"dt":1738724400,"main":{"temp":9.09,"feels_like":7.97,"temp_min":9.09,"temp_max":9.09,"pressure":1033,"sea_level":1033,"grnd_level":1025,"humidity":69,"temp_kf":0},"weather":[{"id":804,"main":"Clouds","description":"overcast clouds","icon":"04n"}],"clouds":{"all":91},"wind":{"speed":2.2,"deg":77,"gust":2.25},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-05 03:00:00"},{"dt":1738735200,"main":{"temp":8.43,"feels_like":7.3,"temp_min":8.43,"temp_max":8.43,"pressure":1033,"sea_level":1033,"grnd_level":1024,"humidity":69,"temp_kf":0},"weather":[{"id":803,"main":"Clouds","description":"broken clouds","icon":"04n"}],"clouds":{"all":58},"wind":{"speed":2.07,"deg":88,"gust":2.2},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-05 06:00:00"},{"dt":1738746000,"main":{"temp":8.87,"feels_like":8.07,"temp_min":8.87,"temp_max":8.87,"pressure":1035,"sea_level":1035,"grnd_level":1026,"humidity":67,"temp_kf":0},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":7},"wind":{"speed":1.78,"deg":85,"gust":2.19},"visibility":10000,"pop":0,"sys":{"pod":"d"},"dt_txt":"2025-02-05 09:00:00"},{"dt":1738756800,"main":{"temp":13.04,"feels_like":11.73,"temp_min":13.04,"temp_max":13.04,"pressure":1035,"sea_level":1035,"grnd_level":1026,"humidity":51,"temp_kf":0},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":7},"wind":{"speed":1.61,"deg":90,"gust":2.52},"visibility":10000,"pop":0,"sys":{"pod":"d"},"dt_txt":"2025-02-05 12:00:00"},{"dt":1738767600,"main":{"temp":14.58,"feels_like":13.32,"temp_min":14.58,"temp_max":14.58,"pressure":1032,"sea_level":1032,"grnd_level":1024,"humidity":47,"temp_kf":0},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":10},"wind":{"speed":1.17,"deg":259,"gust":1.98},"visibility":10000,"pop":0,"sys":{"pod":"d"},"dt_txt":"2025-02-05 15:00:00"},{"dt":1738778400,"main":{"temp":12.58,"feels_like":11.49,"temp_min":12.58,"temp_max":12.58,"pressure":1033,"sea_level":1033,"grnd_level":1024,"humidity":61,"temp_kf":0},"weather":[{"id":801,"main":"Clouds","description":"few clouds","icon":"02n"}],"clouds":{"all":20},"wind":{"speed":2.21,"deg":331,"gust":1.99},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-05 18:00:00"},{"dt":1738789200,"main":{"temp":11.6,"feels_like":10.51,"temp_min":11.6,"temp_max":11.6,"pressure":1033,"sea_level":1033,"grnd_level":1025,"humidity":65,"temp_kf":0},"weather":[{"id":803,"main":"Clouds","description":"broken clouds","icon":"04n"}],"clouds":{"all":79},"wind":{"speed":1.71,"deg":43,"gust":1.89},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-05 21:00:00"},{"dt":1738800000,"main":{"temp":10.55,"feels_like":9.23,"temp_min":10.55,"temp_max":10.55,"pressure":1033,"sea_level":1033,"grnd_level":1024,"humidity":60,"temp_kf":0},"weather":[{"id":804,"main":"Clouds","description":"overcast clouds","icon":"04n"}],"clouds":{"all":89},"wind":{"speed":2.34,"deg":102,"gust":2.29},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-06 00:00:00"},{"dt":1738810800,"main":{"temp":9.51,"feels_like":8.79,"temp_min":9.51,"temp_max":9.51,"pressure":1032,"sea_level":1032,"grnd_level":1023,"humidity":60,"temp_kf":0},"weather":[{"id":804,"main":"Clouds","description":"overcast clouds","icon":"04n"}],"clouds":{"all":100},"wind":{"speed":1.81,"deg":100,"gust":1.75},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-06 03:00:00"},{"dt":1738821600,"main":{"temp":9.04,"feels_like":8.31,"temp_min":9.04,"temp_max":9.04,"pressure":1031,"sea_level":1031,"grnd_level":1022,"humidity":57,"temp_kf":0},"weather":[{"id":804,"main":"Clouds","description":"overcast clouds","icon":"04n"}],"clouds":{"all":100},"wind":{"speed":1.74,"deg":88,"gust":1.6},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2025-02-06 06:00:00"}],"city":{"id":6458924,"name":"Porto Municipality","coord":{"lat":41.1579,"lon":-8.6291},"country":"PT","population":0,"timezone":0,"sunrise":1738654944,"sunset":1738691669}}`
	weatherData := &WeatherData{}
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return nil, err
	}
	return weatherData, nil
}
