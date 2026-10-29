package weather

import (
	"fmt"
	"math"
	"time"
)

// helpers that prepare the forecast for the templates

const (
	hourlySlots = 8
	// probability of precipitation from which a slot counts as rainy
	popRainy = 0.5
)

type DayForecast struct {
	// Today, Tomorrow or the weekday
	Label       string
	Date        time.Time
	Emoji       string
	Description string
	Min         float64
	Max         float64
	Rain        float64
	MaxPop      float64
	MaxWind     float64

	// position of the min/max bar relative to the range of the whole forecast, in percent
	BarLeft  int
	BarWidth int
}

// Current returns the forecast closest to now.
func (w *WeatherData) Current() *WeatherEntry {
	if len(w.List) == 0 {
		return nil
	}
	return w.List[0]
}

// Hourly returns the forecasts of the next 24 hours.
func (w *WeatherData) Hourly() []*WeatherEntry {
	if len(w.List) <= hourlySlots {
		return w.List
	}
	return w.List[:hourlySlots]
}

func (e *WeatherEntry) IsRainy() bool {
	return e.Rain.H3 >= rainLight || e.Pop >= popRainy
}

// RainOutlook describes when rain starts or stops within the next 24 hours.
func (w *WeatherData) RainOutlook() string {
	hourly := w.Hourly()
	if len(hourly) == 0 {
		return ""
	}

	if hourly[0].IsRainy() {
		for _, entry := range hourly[1:] {
			if !entry.IsRainy() {
				return fmt.Sprintf("Rain until %s", entry.Time.Format("15:04"))
			}
		}
		return "Rain all day"
	}

	for _, entry := range hourly[1:] {
		if entry.IsRainy() {
			return fmt.Sprintf("Rain from %s", entry.Time.Format("15:04"))
		}
	}
	return "Dry next 24h"
}

// Days aggregates the forecast per day.
func (w *WeatherData) Days() []DayForecast {
	var days []DayForecast
	var emojiFreq map[string]int

	for _, entry := range w.List {
		date := entry.Time.Format("2006-01-02")
		if len(days) == 0 || days[len(days)-1].Date.Format("2006-01-02") != date {
			label := entry.Time.Weekday().String()[:3]
			switch date {
			case w.Now:
				label = "Today"
			case w.Tomorrow:
				label = "Tomorrow"
			}
			days = append(days, DayForecast{
				Label: label,
				Date:  entry.Time,
				Min:   entry.Main.Temp,
				Max:   entry.Main.Temp,
			})
			emojiFreq = map[string]int{}
		}

		day := &days[len(days)-1]
		day.Min = math.Min(day.Min, entry.Main.Temp)
		day.Max = math.Max(day.Max, entry.Main.Temp)
		day.Rain += entry.Rain.H3
		day.MaxPop = math.Max(day.MaxPop, entry.Pop)
		day.MaxWind = math.Max(day.MaxWind, entry.Wind.Speed)

		// the most frequent condition during daytime describes the day best
		weight := 1
		if entry.Time.Hour() >= 9 && entry.Time.Hour() <= 18 {
			weight = 3
		}
		emojiFreq[entry.WeatherEmoji] += weight
		if emojiFreq[entry.WeatherEmoji] > emojiFreq[day.Emoji] || day.Emoji == "" {
			day.Emoji = entry.WeatherEmoji
			day.Description = entry.WeatherDescription
		}
	}

	if len(days) == 0 {
		return nil
	}

	lowest, highest := days[0].Min, days[0].Max
	for _, day := range days {
		lowest = math.Min(lowest, day.Min)
		highest = math.Max(highest, day.Max)
	}
	span := highest - lowest
	for i := range days {
		if span <= 0 {
			days[i].BarWidth = 100
			continue
		}
		days[i].BarLeft = int(math.Round((days[i].Min - lowest) / span * 100))
		// keep a minimum width so a day with a constant temperature is still visible
		days[i].BarWidth = max(4, int(math.Round((days[i].Max-days[i].Min)/span*100)))
		if days[i].BarLeft+days[i].BarWidth > 100 {
			days[i].BarLeft = 100 - days[i].BarWidth
		}
	}

	return days
}
