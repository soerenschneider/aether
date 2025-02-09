package weather

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type SlotData struct {
	descriptions map[string]struct{}
	tempSum      float64
	count        int
	wind         float64
	precip       float64
	emojiFreq    map[string]int
}

func GenerateWeatherReport(entries []*WeatherEntry, currentTime time.Time) []string {
	startOfDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	endOfNight := startOfDay.Add(29 * time.Hour) // Includes next day until 5 AM

	weatherSlots := map[string]*SlotData{
		"Morning":   {descriptions: make(map[string]struct{}), emojiFreq: make(map[string]int)},
		"Afternoon": {descriptions: make(map[string]struct{}), emojiFreq: make(map[string]int)},
		"Evening":   {descriptions: make(map[string]struct{}), emojiFreq: make(map[string]int)},
		"Night":     {descriptions: make(map[string]struct{}), emojiFreq: make(map[string]int)},
	}

	for _, entry := range entries {
		if entry.Time.Before(startOfDay) || entry.Time.After(endOfNight) {
			continue // Skip entries that are not in today's range
		}

		hour := entry.Time.Hour()
		slot := "Night"
		switch {
		case hour >= 6 && hour < 12:
			slot = "Morning"
		case hour >= 12 && hour < 17:
			slot = "Afternoon"
		case hour >= 17 && hour < 22:
			slot = "Evening"
		case hour >= 22 || hour < 6:
			slot = "Night"
		}

		slotData := weatherSlots[slot]
		slotData.descriptions[entry.WeatherDescription] = struct{}{}
		slotData.tempSum += (entry.Main.TempMax + entry.Main.TempMin) / 2
		slotData.count++
		if entry.Wind.Speed >= WindCalm {
			slotData.wind = entry.Wind.Speed
		}
		if entry.Pop > rainLight {
			slotData.precip = entry.Pop
		}

		// Count the frequency of each WeatherEmoji
		if entry.WeatherEmoji != "" {
			slotData.emojiFreq[entry.WeatherEmoji]++
		}
	}

	timeSlots := []string{"Morning", "Afternoon", "Evening", "Night"}
	var reports []string
	for _, slot := range timeSlots {
		slotData := weatherSlots[slot]
		if slotData.count == 0 {
			continue
		}

		var descriptions []string
		for desc := range slotData.descriptions {
			if strings.Contains(strings.ToLower(desc), "rain") {
				emoji := getEmojiForPrecipitation(slotData.precip)
				amount := formatRainAmount(slotData.precip)
				desc = fmt.Sprintf("%s %s %s", desc, emoji, amount)
			}
			descriptions = append(descriptions, desc)
		}
		sort.Strings(descriptions)

		var emojis []string
		for emoji := range slotData.emojiFreq {
			emojis = append(emojis, emoji)
		}
		sort.Strings(emojis) // Sort for consistent order
		emojiString := strings.Join(emojis, " ")

		avgTemp := slotData.tempSum / float64(slotData.count)
		report := fmt.Sprintf("%s %s will have %s with an avg. temp of %.0f°C",
			emojiString, slot, strings.Join(descriptions, " and "), avgTemp)

		// this should probably never happen. if the description doesn't mention rain but there's
		// precipitation, we add it to the report.
		if slotData.precip >= rainLight && !ContainsRain(report) {
			amount := formatRainAmount(slotData.precip)
			report += fmt.Sprintf(", and %s %s", convertPrecipitation(slotData.precip), amount)
		}
		if slotData.wind >= WindCalm {
			report += fmt.Sprintf(", with %s (%.0f m/s)", WindSpeedDescription(slotData.wind), slotData.wind)
		}
		report += "."

		reports = append(reports, report)
	}

	return reports
}

func formatRainAmount(precip float64) string {

	amount := fmt.Sprintf("(%.0f mm)", precip)
	if precip < 1 {
		amount = fmt.Sprintf("(%.1f mm)", precip)
	}
	return amount
}

func ContainsRain(report string) bool {
	return strings.Contains(strings.ToLower(report), "rain")
}

// getDominantEmoji finds the most frequently occurring emoji in a given slot.
func getDominantEmoji(emojiFreq map[string]int) string {
	var dominantEmoji string
	maxCount := 0
	for emoji, count := range emojiFreq {
		if count > maxCount {
			dominantEmoji = emoji
			maxCount = count
		}
	}
	return dominantEmoji
}
