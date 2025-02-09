package weather

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func getDummyWeather(niceName string) (*WeatherData, error) {
	body, err := os.ReadFile(fmt.Sprintf("../../../contrib/testdata/weather-%s.json", strings.ToLower(niceName)))
	if err != nil {
		return nil, err
	}
	weatherData := &WeatherData{}
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return nil, err
	}

	return weatherData, nil
}
