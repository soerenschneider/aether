package cached

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/soerenschneider/aether/internal/datasource"

	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

const defaultRefreshInterval = 5 * time.Minute

type Opts func(cachedDatasource *CachedDatasource) error

func New(datasource datasource.Datasource, opts ...Opts) (*CachedDatasource, error) {
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
	html string

	mutex           sync.RWMutex
	nextRefresh     time.Time
	refreshInterval time.Duration
	datasource      datasource.Datasource
}

func (b *CachedDatasource) GetHtml(ctx context.Context) (string, error) {
	if b.nextRefresh.IsZero() || time.Now().After(b.nextRefresh) {
		b.mutex.Lock()
		defer b.mutex.Unlock()

		if b.nextRefresh.IsZero() || time.Now().After(b.nextRefresh) {
			log.Debug().Msgf("Updating cached datasource %q", b.datasource.Name())
			html, err := b.datasource.GetHtml(ctx)
			if err == nil {
				b.html = html
				b.nextRefresh = time.Now().Add(b.refreshInterval)
				return html, nil
			}

			if len(b.html) > 0 {
				log.Error().Err(err).Msg("got error")
				return b.html, nil
			}
			return "", err
		}

		return b.html, nil
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.html, nil
}

func (b *CachedDatasource) Name() string {
	return fmt.Sprintf("Cached %s", b.datasource.Name())
}
