package datasource

import "context"

type Datasource interface {
	GetHtml(ctx context.Context) (string, error)
	Name() string
}
