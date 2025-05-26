package logs

import (
	"net/url"
	"path"
)

type LogEntry struct {
	unix      int64
	Timestamp string `json:"_time"`
	Message   string `json:"_msg"`
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
