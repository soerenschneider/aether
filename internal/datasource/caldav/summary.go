package caldav

import (
	"fmt"
	"strings"
	"time"

	"github.com/soerenschneider/aether/pkg"
)

func getSummary(entries []Entry, now time.Time, addSummaryForNoEvents bool) []string {
	var ret []string
	for _, entry := range entries {
		isToday := pkg.IsToday(entry.Start, now)
		isOngoing := entry.Start.Before(now) && entry.End.After(now)

		if isToday || isOngoing {
			b := fmt.Sprintf("📅 %s, %s", entry.Summary, strings.Join(entry.Formatted, " "))
			if entry.Calendar != "" {
				b += fmt.Sprintf(" (%s)", entry.Calendar)
			}
			ret = append(ret, b)
		}
	}

	if addSummaryForNoEvents && len(ret) == 0 {
		ret = append(ret, "✅ No events for today")
	}

	return ret
}
