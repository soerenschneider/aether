package caldav

import (
	"errors"
	"fmt"
	"html/template"
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

// WithMonthTemplate enables the month view of the calendar, see PageName.
func WithMonthTemplate(data []byte) DatasourceOpt {
	return func(ds *CaldavDatasource) error {
		tmpl, err := template.New("calendar-month").Parse(string(data))
		if err != nil {
			return fmt.Errorf("could not parse month template: %w", err)
		}
		ds.monthTemplate = tmpl
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
