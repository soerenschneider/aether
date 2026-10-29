package logs

import (
	"net/url"
	"path"
	"time"
)

type LogEntry struct {
	unix      int64
	Timestamp string `json:"_time"`
	Message   string `json:"_msg"`
}

// formatClock shortens the RFC3339 timestamps of victorialogs to the local time.
func formatClock(timestamp string) string {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return timestamp
	}

	t = t.Local()
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04:05")
	}
	return t.Format("02.01. 15:04")
}

type VictorialogsQuery struct {
	Address string
	Query   string
	Limit   int
}

func (q *VictorialogsQuery) GetLimit() int {
	if q.Limit <= 0 || q.Limit > 500 {
		return 25
	}

	return q.Limit
}

func buildURL(baseAddr, endpointPath string) (string, error) {
	u, err := url.Parse(baseAddr)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, endpointPath)

	return u.String(), nil
}
