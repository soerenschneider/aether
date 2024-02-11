package static

import (
	"context"
	"sync"
)

func NewStatic(html string) *StaticDatasource {
	return &StaticDatasource{
		html: html,
	}
}

type StaticDatasource struct {
	html  string
	mutex sync.RWMutex
}

func (b *StaticDatasource) Update(html string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.html = html
}

func (b *StaticDatasource) Name() string {
	return "static"
}

func (b *StaticDatasource) GetHtml(_ context.Context) (string, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.html, nil
}
