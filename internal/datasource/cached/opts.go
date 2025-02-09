package cached

import (
	"errors"
	"time"
)

func WithRefreshInterval(interval time.Duration) Opts {
	return func(ds *CachedDatasource) error {
		if interval < 1*time.Minute {
			return errors.New("min interval is 1 minute")
		}

		ds.refreshInterval = interval
		return nil
	}
}
