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

	HtmlId string
}

type Card struct {
	Name          string
	Anniversary   time.Time
	Upcoming      time.Time
	DateFormatted string
	Type          string
	TypeEmoji     string
	Years         int
}

// Today returns the cards whose anniversary is today.
func (d CarddavData) Today() []Card {
	var ret []Card
	for _, card := range d.Cards {
		if card.IsToday() {
			ret = append(ret, card)
		}
	}
	return ret
}

func (c Card) IsToday() bool {
	return pkg.IsToday(c.Upcoming, time.Now())
}

// When describes the next occurrence of the anniversary, e.g. "Today", "Tomorrow" or "Mon 02.11.".
func (c Card) When() string {
	now := time.Now()
	if pkg.IsToday(c.Upcoming, now) {
		return "Today"
	}
	if pkg.IsToday(c.Upcoming, now.AddDate(0, 0, 1)) {
		return "Tomorrow"
	}
	return c.Upcoming.Format("Mon 02.01.")
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

func getFormattedAnniversaryDate(anniversary time.Time, now time.Time) string {
	year := now.Year()
	if anniversary.Month() < now.Month() {
		year += 1
	}

	check := time.Date(year, anniversary.Month(), anniversary.Day(), 12, 0, 0, 0, time.UTC)

	var prefix string
	if pkg.IsToday(check, now) {
		prefix = "Today"
	} else if pkg.IsTomorrow(check) {
		prefix = "Tomorrow"
	} else {
		prefix = check.Weekday().String()[:3]
	}

	f := anniversary.Format("02.01.2006")
	return fmt.Sprintf("%s, %s", prefix, f)
}
