package weather

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

func getClassForTemp(temp float64) string {
	if temp >= 30 {
		return "red"
	}
	if temp >= 25 {
		return "orange"
	}
	if temp < 10 {
		return "lightblue"
	}
	if temp < 20 {
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
	if speed >= 13 {
		return "red"
	}
	if speed > 7.9 {
		return "orange"
	}
	if speed > 3.3 {
		return "yellow"
	}
	return ""
}

func getClassForRain(rain float64) string {
	if rain >= 8 {
		return "red"
	}
	if rain >= 4 {
		return "orange"
	}
	if rain >= 2 {
		return "yellow"
	}
	if rain >= 0.33 {
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
