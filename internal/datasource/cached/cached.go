package cached

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/pkg"

	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

const (
	defaultRefreshInterval = 5 * time.Minute
	// incomplete data is fetched again after this interval, instead of being kept until the regular refresh
	incompleteRetryInterval = time.Minute
)

type Opts func(cachedDatasource *CachedDatasource) error

type Datasource interface {
	GetData(ctx context.Context) (*internal.Data, error)
	Name() string
}

func New(datasource Datasource, opts ...Opts) (*CachedDatasource, error) {
	if datasource == nil {
		return nil, errors.New("cache: empty datasource provided")
	}

	ds := &CachedDatasource{
		datasource:      datasource,
		refreshInterval: defaultRefreshInterval,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(ds); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return ds, errs
}

type CachedDatasource struct {
	data *internal.Data

	mutex           sync.RWMutex
	nextRefresh     time.Time
	refreshInterval time.Duration
	datasource      Datasource
}

func calculateNextRefreshInterval(refreshInterval time.Duration, now time.Time) time.Time {
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	nextRegularRefreshInterval := now.Add(refreshInterval)
	if nextRegularRefreshInterval.Before(nextMidnight) {
		return nextRegularRefreshInterval
	}
	return nextMidnight
}

func (b *CachedDatasource) GetData(ctx context.Context) (*internal.Data, error) {
	if b.nextRefresh.IsZero() || time.Now().After(b.nextRefresh) {
		b.mutex.Lock()
		defer b.mutex.Unlock()

		if b.nextRefresh.IsZero() || time.Now().After(b.nextRefresh) {
			log.Debug().Msgf("Updating cached datasource %q", b.datasource.Name())
			data, err := b.datasource.GetData(ctx)
			if err == nil {
				if data != nil {
					data.FetchedAt = time.Now()
				}
				b.data = data
				b.nextRefresh = calculateNextRefreshInterval(b.refreshInterval, time.Now())
				if data != nil && data.Incomplete {
					b.nextRefresh = calculateNextRefreshInterval(min(b.refreshInterval, incompleteRetryInterval), time.Now())
					log.Debug().Msgf("Data of cached datasource %q is incomplete, retrying at %v", b.datasource.Name(), b.nextRefresh)
				}
				return data, nil
			}

			// show the last data instead, the next call tries again. The rendered data contains relative dates
			// ("today"), so it's only usable on the day it was fetched.
			if b.data != nil && pkg.IsToday(b.data.FetchedAt, time.Now()) {
				log.Warn().Err(err).Msgf("Could not update cached datasource %q, showing data from %v", b.datasource.Name(), b.data.FetchedAt)
				outdated := *b.data
				outdated.Outdated = true
				return &outdated, nil
			}

			return nil, err
		}

		return b.data, nil
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.data, nil
}

// Unwrap returns the datasource that is cached.
func (b *CachedDatasource) Unwrap() Datasource {
	return b.datasource
}

func (b *CachedDatasource) Name() string {
	return b.datasource.Name()
}
