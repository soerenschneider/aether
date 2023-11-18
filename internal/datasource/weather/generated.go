package weather

import (
	"encoding/json"
	"time"
)

type WeatherData struct {
	Cod     string         `json:"cod"`
	Message int            `json:"message"`
	Cnt     int            `json:"cnt"`
	List    []WeatherEntry `json:"list"`
	City    City           `json:"city"`
}

type WeatherEntry struct {
	Dt         int       `json:"dt"`
	Main       Main      `json:"main"`
	Weather    []Weather `json:"weather"`
	Clouds     Clouds    `json:"clouds"`
	Wind       Wind      `json:"wind"`
	Visibility int       `json:"visibility"`
	Pop        float64   `json:"pop"`
	Rain       Rain      `json:"rain"`
	Sys        Sys       `json:"sys"`
	DtTxt      string    `json:"dt_txt"`

	Time               time.Time
	WeatherDescription string
	WeatherIconName    string
	WeatherEmoji       string
	WeatherLink        string
	PopClass           string
	VisibilityPercent  int
	VisibilityCssClass string
}

func (w *WeatherEntry) UnmarshalJSON(data []byte) error {
	type Alias WeatherEntry

	tmp := Alias{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	tmp.Time = time.Unix(int64(tmp.Dt), 0)
	tmp.PopClass = getClassForPop(tmp.Pop)
	if len(tmp.Weather) > 0 {
		tmp.WeatherDescription = tmp.Weather[0].Description
		tmp.WeatherEmoji = getIconByWeatherId(tmp.Weather[0].ID)
		tmp.WeatherIconName = tmp.Weather[0].Icon
	}
	tmp.VisibilityPercent = tmp.Visibility * 100 / 10000
	tmp.VisibilityCssClass = getClassForVisibility(tmp.VisibilityPercent)

	*w = WeatherEntry(tmp)
	return nil
}

type Main struct {
	Temp          float64 `json:"temp"`
	TempClass     string
	FeelsLike     float64 `json:"feels_like"`
	TempMin       float64 `json:"temp_min"`
	TempMax       float64 `json:"temp_max"`
	Pressure      int     `json:"pressure"`
	SeaLevel      int     `json:"sea_level"`
	GrndLevel     int     `json:"grnd_level"`
	Humidity      int     `json:"humidity"`
	HumidityClass string
	TempKf        float64 `json:"temp_kf"`
}

func (w *Main) UnmarshalJSON(data []byte) error {
	type Alias Main

	tmp := Alias{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	// Use the feelsLike temperature to calculate the class
	tmp.TempClass = getClassForTemp(tmp.FeelsLike)
	tmp.HumidityClass = getClassForHumidity(tmp.Humidity)

	*w = Main(tmp)
	return nil
}

type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func (w *Clouds) UnmarshalJSON(data []byte) error {
	type Alias Clouds

	tmp := Alias{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	tmp.CssClass = getClassForClouds(tmp.All)

	*w = Clouds(tmp)
	return nil
}

type Clouds struct {
	All      int `json:"all"`
	CssClass string
}

func (w *Wind) UnmarshalJSON(data []byte) error {
	type Alias Wind

	tmp := Alias{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	tmp.CssClass = getClassForWind(tmp.Speed)

	*w = Wind(tmp)
	return nil
}

type Wind struct {
	Speed    float64 `json:"speed"`
	CssClass string
	Deg      int     `json:"deg"`
	Gust     float64 `json:"gust"`
}

type Rain struct {
	H3       float64 `json:"3h"`
	CssClass string
}

func (w *Rain) UnmarshalJSON(data []byte) error {
	type Alias Rain

	tmp := Alias{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	tmp.CssClass = getClassForRain(tmp.H3)

	*w = Rain(tmp)
	return nil
}

type Sys struct {
	Pod string `json:"pod"`
}

type City struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Coord       Coord  `json:"coord"`
	Country     string `json:"country"`
	Population  int    `json:"population"`
	Timezone    int    `json:"timezone"`
	Sunrise     int    `json:"sunrise"`
	SunriseTime time.Time
	Sunset      int `json:"sunset"`
	SunsetTime  time.Time
}

func (w *City) UnmarshalJSON(data []byte) error {
	type Alias City

	tmp := Alias{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	tmp.SunriseTime = time.Unix(int64(tmp.Sunrise), 0)
	tmp.SunsetTime = time.Unix(int64(tmp.Sunset), 0)

	*w = City(tmp)
	return nil
}

type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
