package weather

import "fmt"

func getClassForPop(pop float64) string {
	if pop > 0.75 {
		return "red"
	}
	if pop > 0.5 {
		return "orange"
	}
	if pop > 0.33 {
		return "yellow"
	}
	return ""
}

const (
	tempVeryHot  = 35
	tempHot      = 30
	tempWarm     = 25
	tempMedium   = 20
	tempCold     = 10
	tempVeryCold = 0
)

func getClassForTemp(temp float64) string {
	if temp >= tempHot {
		return "red"
	}
	if temp >= tempWarm {
		return "orange"
	}
	if temp < tempCold {
		return "lightblue"
	}
	if temp < tempMedium {
		return "blue"
	}
	return ""
}

func getClassForHumidity(humidity int) string {
	if humidity >= 90 {
		return "red"
	}
	if humidity > 70 {
		return "orange"
	}
	if humidity < 30 {
		return "yellow"
	}
	return ""
}

func getClassForClouds(all int) string {
	if all > 75 {
		return "red"
	}
	if all > 50 {
		return "orange"
	}
	if all > 25 {
		return "yellow"
	}

	return ""
}

func getClassForVisibility(all int) string {
	if all > 75 {
		return ""
	}
	if all > 50 {
		return "yellow"
	}
	if all > 25 {
		return "orange"
	}

	return "red"
}

func getClassForWind(speed float64) string {
	if speed >= WindGale {
		return "red"
	}
	if speed >= WindStrongBreeze {
		return "orange"
	}
	if speed >= WindModerateBreeze {
		return "yellow"
	}
	return ""
}

const (
	rainExtreme   = 50
	rainVeryHeavy = 20
	rainHeavy     = 7.6
	rainModerate  = 2.6
	rainLight     = 0.1
)

func getEmojiForPrecipitation(rain3H float64) string {
	if rain3H >= rainExtreme {
		return "🌊"
	}
	if rain3H >= rainVeryHeavy {
		return "⛈️"
	}
	if rain3H >= rainHeavy {
		return "🌧️🌧️"
	}
	if rain3H >= rainModerate {
		return "🌧️"
	}
	if rain3H >= rainLight {
		return "🌦️"
	}
	return ""
}

func convertPrecipitation(rain3H float64) string {
	if rain3H >= rainExtreme {
		return fmt.Sprintf("extreme rain %s", getEmojiForPrecipitation(rain3H))
	}
	if rain3H >= rainVeryHeavy {
		return fmt.Sprintf("very heavy rain️ %s", getEmojiForPrecipitation(rain3H))
	}
	if rain3H >= rainHeavy {
		return fmt.Sprintf("heavy rain️ %s", getEmojiForPrecipitation(rain3H))
	}
	if rain3H >= rainModerate {
		return fmt.Sprintf("moderate rain %s", getEmojiForPrecipitation(rain3H))
	}
	if rain3H >= rainLight {
		return fmt.Sprintf("light rain %s", getEmojiForPrecipitation(rain3H))
	}
	return ""
}

func GetPrecipitationDescription(rain3H float64) string {
	if rain3H >= rainExtreme {
		return "Torrential rain, possible flash flooding."
	}
	if rain3H >= rainVeryHeavy {
		return "Very heavy rain, risk of localized flooding."
	}
	if rain3H >= rainHeavy {
		return "Heavy rain, roads may become wet and slippery."
	}
	if rain3H >= rainModerate {
		return "Moderate rain, occasional showers expected."
	}
	if rain3H >= rainLight {
		return "Light rain, a few drizzles throughout the period."
	}
	return "No rain, dry weather expected."
}

func getClassForRain(rain float64) string {
	if rain >= rainExtreme {
		return "red"
	}
	if rain >= rainHeavy {
		return "orange"
	}
	if rain >= rainModerate {
		return "yellow"
	}
	if rain >= rainLight {
		return "blue"
	}
	return ""
}

func getIconByWeatherId(id int) string {
	if id > 800 {
		return "☁️"
	}
	if id == 800 {
		return "☀️"
	}
	if id > 700 {
		return "🌪"
	}
	if id > 600 {
		return "❄️"
	}
	if id > 500 {
		return "☔️"
	}
	if id > 300 {
		return "☂️"
	}
	if id > 200 {
		return "⚡️"
	}

	return ""
}

const (
	WindCalm           = 1
	WindLightBreeze    = 3
	WindModerateBreeze = 6
	WindStrongBreeze   = 10
	WindGale           = 15
	WindStrongGale     = 20
	WindStorm          = 30
)

// WindSpeedEmoji returns the emoji based on wind speed
func WindSpeedEmoji(windSpeed float64) string {
	switch {
	case windSpeed >= WindStorm:
		return "💨"
	case windSpeed >= WindStrongGale:
		return "🌬️"
	case windSpeed >= WindGale:
		return "🌪️"
	case windSpeed >= WindStrongBreeze:
		return "🌬️"
	case windSpeed >= WindModerateBreeze:
		return "🍃"
	case windSpeed >= WindLightBreeze:
		return "🌿"
	case windSpeed >= WindCalm:
		return "🐢"
	}
	return ""
}

// WindSpeedDescription returns a description based on wind speed
func WindSpeedDescription(windSpeed float64) string {
	switch {
	case windSpeed >= WindStorm:
		return fmt.Sprintf("very strong winds %s", WindSpeedEmoji(windSpeed))
	case windSpeed >= WindStrongGale:
		return fmt.Sprintf("strong gale %s", WindSpeedEmoji(windSpeed))
	case windSpeed >= WindGale:
		return fmt.Sprintf("gale %s", WindSpeedEmoji(windSpeed))
	case windSpeed >= WindStrongBreeze:
		return fmt.Sprintf("strong breeze %s", WindSpeedEmoji(windSpeed))
	case windSpeed >= WindModerateBreeze:
		return fmt.Sprintf("breeze %s", WindSpeedEmoji(windSpeed))
	case windSpeed >= WindLightBreeze:
		return fmt.Sprintf("light breeze %s", WindSpeedEmoji(windSpeed))
	case windSpeed >= WindCalm:
		return fmt.Sprintf("calm wind %s", WindSpeedEmoji(windSpeed))
	}
	return ""
}
