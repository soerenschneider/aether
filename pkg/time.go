package pkg

import (
	"fmt"
	"time"
)

func Today(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func OneWeek(now time.Time) time.Time {
	return NWeeks(now, 7)
}

func TwoWeeks(now time.Time) time.Time {
	return NWeeks(now, 14)
}

func NWeeks(now time.Time, days int) time.Time {
	end := now.AddDate(0, 0, days)
	return time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
}

func AtSameDay(a time.Time, b time.Time) bool {
	if a.Year() != b.Year() {
		return false
	}

	if a.Month() != b.Month() {
		return false
	}

	if a.Day() == b.Day() {
		return true
	}

	diff := b.Sub(a)
	if b.Hour() == 0 && diff.Hours() >= 0 && diff.Hours() <= 24 {
		return true
	}

	return false
}

func IsToday(date time.Time, now time.Time) bool {
	if date.Year() != now.Year() {
		return false
	}
	if date.Month() != now.Month() {
		return false
	}

	return date.Day() == now.Day()
}

func IsTomorrow(date time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	until := time.Until(date)
	now := time.Now()

	return until > 0 && until <= 24*time.Hour && now.Day() != date.Day()
}

func StartOfWeek(now time.Time) time.Time {
	daysUntilMonday := int(time.Monday - now.Weekday())
	if daysUntilMonday < 0 {
		daysUntilMonday += 7
	}
	start := now.AddDate(0, 0, -daysUntilMonday)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	return start
}

func EndOfWeek(now time.Time) time.Time {
	daysUntilSunday := int(time.Sunday - now.Weekday())
	if daysUntilSunday < 0 {
		daysUntilSunday += 7
	}
	end := now.AddDate(0, 0, daysUntilSunday)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
	return end
}

func DurationToString(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	years := days / 365
	months := (days % 365) / 30
	remainingDays := (days % 365) % 30

	var result string

	if years > 0 {
		result += fmt.Sprintf("%d years, ", years)
	}

	if months > 0 {
		result += fmt.Sprintf("%d months, ", months)
	}

	if remainingDays > 0 || (years == 0 && months == 0) {
		result += fmt.Sprintf("%d days", remainingDays)
	}

	return result
}

func IsWholeDay(start, end time.Time) bool {
	if start.Hour() != 0 || end.Hour() != 0 {
		return false
	}

	return start.Minute() == 0 && end.Minute() == 0
}
