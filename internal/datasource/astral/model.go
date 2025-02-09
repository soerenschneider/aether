package astral

import "time"

type AstralData struct {
	BlueHourRising         TimeDuration
	BlueHourSetting        TimeDuration
	AzimuthBlueHourRising  float64
	AzimuthBlueHourSetting float64

	GoldenHourRising         TimeDuration
	GoldenHourSetting        TimeDuration
	AzimuthGoldenHourRising  float64
	AzimuthGoldenHourSetting float64

	Sunrise        time.Time
	Sunset         time.Time
	AzimuthSunrise float64
	AzimuthSunset  float64
}

type TimeDuration struct {
	Start time.Time
	End   time.Time
}
