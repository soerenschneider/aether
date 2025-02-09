package pkg

func TranslateDegreeToDirectionEmoji(degrees float64) string {
	_, emoji := TranslateDegreeToDirection(degrees)
	return emoji
}

func TranslateDegreeToDirection(degrees float64) (string, string) {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S",
		"SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}

	emojis := []string{"⬆️", "↗️", "↗️", "↗️", "➡️", "↘️", "↘️", "↘️",
		"⬇️", "↙️", "↙️", "↙️", "⬅️", "↖️", "↖️", "↖️"}

	index := int((degrees+11.25)/22.5) % 16
	return directions[index], emojis[index]

	//switch {
	//case deg >= 0 && deg < 23:
	//	return "North", "⬆️"
	//case deg >= 23 && deg < 68:
	//	return "Northeast", "↗️"
	//case deg >= 68 && deg < 113:
	//	return "East", "➡️"
	//case deg >= 113 && deg < 158:
	//	return "Southeast", "↘️"
	//case deg >= 158 && deg < 203:
	//	return "South", "⬇️"
	//case deg >= 203 && deg < 248:
	//	return "Southwest", "↙️"
	//case deg >= 248 && deg < 293:
	//	return "West", "⬅️"
	//case deg >= 293 && deg < 338:
	//	return "Northwest", "↖️"
	//}
	//
	//return "North", "⬆️"
}
