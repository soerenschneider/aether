package static

import (
	"context"
	"sync"

	"github.com/soerenschneider/aether/internal"
)

func NewStatic(data *internal.Data) *StaticDatasource {
	return &StaticDatasource{
		data: data,
	}
}

type StaticDatasource struct {
	data  *internal.Data
	mutex sync.RWMutex
}

func (b *StaticDatasource) Update(data *internal.Data) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.data = data
}

func (b *StaticDatasource) Name() string {
	return "static"
}

func (b *StaticDatasource) GetData(_ context.Context) (*internal.Data, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.data, nil
}
