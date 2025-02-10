package astral

import (
	"fmt"
	"strings"
	"time"
)

func getSummary(data AstralData, now time.Time) []string {
	summary := []string{}

	if now.Before(data.Sunset) {
		if !data.Sunrise.IsZero() && !data.Sunset.IsZero() {
			summary = append(summary, fmt.Sprintf("☀️Sun: %s - %s", data.Sunrise.Format("15:04"), data.Sunset.Format("15:04")))
		}
	}

	rising := []string{}
	if now.Before(data.BlueHourRising.End) {
		rising = append(rising, fmt.Sprintf("🌌 Blue Hour ⬆️ %s - %s", data.BlueHourRising.Start.Format("15:04"), data.BlueHourRising.End.Format("15:04")))
	}
	if now.Before(data.GoldenHourRising.End) {
		rising = append(rising, fmt.Sprintf("🌇 Golden Hour ⬆️️ %s - %s", data.GoldenHourRising.Start.Format("15:04"), data.GoldenHourRising.End.Format("15:04")))
	}

	risingStr := strings.Join(rising, ", ")

	setting := []string{}
	if now.Before(data.GoldenHourSetting.End) {
		setting = append(setting, fmt.Sprintf("🌇 Golden Hour ⬇️️ %s - %s", data.GoldenHourSetting.Start.Format("15:04"), data.GoldenHourSetting.End.Format("15:04")))
	}

	if now.Before(data.BlueHourSetting.End) {
		setting = append(setting, fmt.Sprintf("🌌 Blue Hour ⬇️ %s - %s", data.BlueHourSetting.Start.Format("15:04"), data.BlueHourSetting.End.Format("15:04")))
	}

	settingStr := strings.Join(setting, ", ")

	if risingStr != "" {
		summary = append(summary, risingStr)
	}

	if settingStr != "" {
		summary = append(summary, settingStr)
	}

	return summary
}
