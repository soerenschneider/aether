package cached

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/soerenschneider/aether/internal"

	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

const defaultRefreshInterval = 5 * time.Minute

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
				b.data = data
				b.nextRefresh = calculateNextRefreshInterval(b.refreshInterval, time.Now())
				return data, nil
			}

			return nil, err
		}

		return b.data, nil
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.data, nil
}

func (b *CachedDatasource) Name() string {
	return b.datasource.Name()
}
