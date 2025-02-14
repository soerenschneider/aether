package carddav

import (
	"fmt"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

func getSummary(entries []Card, now time.Time, addSummaryForNoEvents bool) []string {
	var ret []string
	for _, entry := range entries {
		anniversary := time.Date(now.Year(), entry.Anniversary.Month(), entry.Anniversary.Day(), 12, 0, 0, 0, time.UTC)
		if pkg.IsToday(anniversary, now) {
			b := fmt.Sprintf("%s %s, %s (%d)", entry.TypeEmoji, entry.Name, entry.Type, entry.Years)
			ret = append(ret, b)
		}
	}

	if addSummaryForNoEvents && len(ret) == 0 {
		ret = append(ret, "✅ No anniversaries or birthdays today")
	}

	return ret
}
