package carddav

import (
	"fmt"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

type CarddavData struct {
	Cards []Card
	From  time.Time
	To    time.Time
}

type Card struct {
	Name          string
	Anniversary   time.Time
	DateFormatted string
	Type          string
	Years         int
}

func parseTimeCard(input string) (time.Time, error) {
	layoutTime := "20060102T150405"
	layoutDay := "2006-01-02"

	parsedTime, err := time.Parse(layoutTime, input)
	if err != nil {
		parsedTime, err = time.Parse(layoutDay, input)
		if err != nil {
			return time.Time{}, err
		}
	}

	return parsedTime, nil
}

func getFormattedAnniversaryDate(date time.Time) string {
	now := time.Now()
	now = time.Date(date.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), date.Location())

	var prefix string
	if pkg.IsToday(date, now) {
		prefix = "Today"
	} else if pkg.IsTomorrow(date) {
		prefix = "Tomorrow"
	} else {
		prefix = date.Weekday().String()[:3]
	}

	f := date.Format("02.01.2006")
	return fmt.Sprintf("%s, %s", prefix, f)
}
