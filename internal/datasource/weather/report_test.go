package weather

import (
	"testing"
	"time"
)

func TestGenerateWeatherReport(t *testing.T) {
	// Use a fixed time to make the test deterministic
	fixedTime := time.Date(2025, time.February, 15, 12, 0, 0, 0, time.UTC)

	entries := []*WeatherEntry{
		// Morning entries
		{
			Time:               fixedTime.Add(time.Hour * -5),
			WeatherDescription: "Sunny",
			WeatherEmoji:       "☀️",
			Main: Main{
				TempMax: 25,
				TempMin: 20,
			},
			Wind: Wind{Speed: 3},
			Pop:  0,
		},
		{
			Time:               fixedTime.Add(time.Hour * -4),
			WeatherDescription: "Clear",
			WeatherEmoji:       "☀️",
			Main: Main{
				TempMax: 24,
				TempMin: 19,
			},
			Wind: Wind{Speed: 2},
			Pop:  0,
		},

		// Afternoon entries
		{
			Time:               fixedTime.Add(time.Hour * 1),
			WeatherDescription: "Partly Cloudy",
			WeatherEmoji:       "⛅",
			Main: Main{
				TempMax: 28,
				TempMin: 23,
			},
			Wind: Wind{Speed: 3},
			Pop:  0,
		},
		{
			Time:               fixedTime.Add(time.Hour * 2),
			WeatherDescription: "Cloudy",
			WeatherEmoji:       "☁️",
			Main: Main{
				TempMax: 27,
				TempMin: 22,
			},
			Wind: Wind{Speed: 5},
			Pop:  1.5,
		},

		// Evening entries
		{
			Time:               fixedTime.Add(time.Hour * 6),
			WeatherDescription: "Rain",
			WeatherEmoji:       "🌧️",
			Main: Main{
				TempMax: 22,
				TempMin: 18,
			},
			Wind: Wind{Speed: 4.5},
			Pop:  3,
		},
		{
			Time:               fixedTime.Add(time.Hour * 7),
			WeatherDescription: "Stormy",
			WeatherEmoji:       "⛈️",
			Main: Main{
				TempMax: 21,
				TempMin: 17,
			},
			Wind: Wind{Speed: 6},
			Pop:  5,
		},

		// Night entries
		{
			Time:               fixedTime.Add(time.Hour * 11),
			WeatherDescription: "Clear",
			WeatherEmoji:       "🌙",
			Main: Main{
				TempMax: 18,
				TempMin: 15,
			},
			Wind: Wind{Speed: 2},
			Pop:  0,
		},
		{
			Time:               fixedTime.Add(time.Hour * 12),
			WeatherDescription: "Cloudy",
			WeatherEmoji:       "☁️",
			Main: Main{
				TempMax: 17,
				TempMin: 14,
			},
			Wind: Wind{Speed: 3},
			Pop:  0,
		},
	}

	// Pass in fixedTime to the function to ensure it works with controlled time
	reports := GenerateWeatherReport(entries, fixedTime)

	expectedReports := []string{
		"☀️ Morning will have Clear and Sunny with an avg. temp of 22°C, with calm wind 🐢 (2 m/s).",
		"⛅ Afternoon will have Cloudy and Partly Cloudy with an avg. temp of 25°C, and light rain 🌦️ (2 mm), with light breeze 🌿 (5 m/s).",
		"🌧️ Evening will have Rain 🌧️ (5 mm) and Stormy with an avg. temp of 20°C, with breeze 🍃 (6 m/s).",
		"🌙 Night will have Clear and Cloudy with an avg. temp of 16°C, with light breeze 🌿 (3 m/s).",
	}

	if len(reports) != len(expectedReports) {
		t.Fatalf("Expected %d reports, got %d", len(expectedReports), len(reports))
	}

	for i, report := range reports {
		if report != expectedReports[i] {
			t.Errorf("Mismatch in report[%d]:\nExpected: %s\nGot: %s", i, expectedReports[i], report)
		}
	}
}
