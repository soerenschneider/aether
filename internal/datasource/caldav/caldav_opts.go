package caldav

import (
	"errors"
	"fmt"
	"time"
)

func WithLocation(location *time.Location) DatasourceOpt {
	return func(ds *CaldavDatasource) error {
		if location == nil {
			return errors.New("empty location")
		}
		ds.location = location
		return nil
	}
}

func WithLookaheadDays(days int) DatasourceOpt {
	return func(ds *CaldavDatasource) error {
		if days < 3 || days > 180 {
			return fmt.Errorf("lookahead days should be [3, 180]")
		}

		ds.maxDays = days
		return nil
	}
}
