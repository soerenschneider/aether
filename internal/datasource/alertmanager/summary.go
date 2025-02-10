package alertmanager

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func getSummary(alerts []*Alert, _ time.Time, addSummaryForNoEvents bool) []string {
	type severityCount struct {
		severity string
		count    int
	}

	severities := map[string]int{}

	for _, alert := range alerts {
		_, found := severities[alert.Severity]
		if !found {
			severities[alert.Severity] = alert.Count
		}
		severities[alert.Severity] = alert.Count + severities[alert.Severity]
	}

	var severityList []severityCount
	totalCnt := 0
	for severity, cnt := range severities {
		totalCnt += cnt
		severityList = append(severityList, severityCount{severity, cnt})
	}

	sort.Slice(severityList, func(i, j int) bool {
		return severityList[i].count > severityList[j].count
	})

	summary := []string{}
	for _, sc := range severityList {
		summary = append(summary, fmt.Sprintf("%d %s", sc.count, sc.severity))
	}

	if len(summary) == 0 && addSummaryForNoEvents {
		return []string{"✅ No active alerts"}
	}

	return []string{fmt.Sprintf("🔥 %d active alerts: %s", totalCnt, strings.Join(summary, ", "))}
}
