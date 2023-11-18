package stocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/soerenschneider/aether/pkg"

	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

const (
	baseUrl         = "https://query1.finance.yahoo.com"
	apiPrefix       = "v8/finance"
	linkBasePath    = "https://finance.yahoo.com/quote"
	defaultInterval = "3mo"
	defaultRange    = "2y"
)

type StocksDatasource struct {
	symbols []string

	template   *template.Template
	once       sync.Once
	httpClient *http.Client
}

type Opts func(datasource *StocksDatasource) error

func New(symbols []string, opts ...Opts) (*StocksDatasource, error) {
	if len(symbols) == 0 {
		return nil, errors.New("no symbols supplied")
	}

	ds := &StocksDatasource{
		symbols:    symbols,
		httpClient: http.DefaultClient,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(ds); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return ds, errs
}

func (s *StocksDatasource) Name() string {
	return "stocks"
}

func (s *StocksDatasource) GetHtml(ctx context.Context) (string, error) {
	s.once.Do(func() {
		if s.template == nil {
			s.template = template.Must(template.New("stocks").Parse(defaultTemplate))
		}
	})

	data, err := s.getStocks(ctx)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = s.template.Execute(&tpl, data)
	return tpl.String(), err
}

func buildChartURL(symbol, interval, timeRange string) (string, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	u.Path = fmt.Sprintf("%s/chart/%s", apiPrefix, symbol)

	q := u.Query()
	q.Set("metrics", "high")
	q.Set("interval", interval)
	q.Set("range", timeRange)

	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (s *StocksDatasource) getStock(ctx context.Context, symbol, interval, timeRange string) (*Response, error) {
	url, err := buildChartURL(symbol, interval, timeRange)
	old := fmt.Sprintf("%s/chart/%s?metrics=high?&interval=%s&range=%s", baseUrl, symbol, interval, timeRange)
	if err != nil {
		return nil, errors.New(old)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	chart := &Response{}
	if err := json.Unmarshal(data, &chart); err != nil {
		return nil, err
	}

	return chart, nil
}

func (s *StocksDatasource) getStocks(ctx context.Context) (*StockData, error) {
	ret := make([]*Response, len(s.symbols))
	errs := make(chan error, len(s.symbols))

	wg := sync.WaitGroup{}
	wg.Add(len(s.symbols))
	for index, symbol := range s.symbols {
		go func(index int, sym string) {
			res, err := s.getStock(ctx, sym, defaultInterval, defaultRange)
			if err != nil {
				log.Error().Err(err).Msgf("could not resolve symbol %q", sym)
				errs <- err
			} else {
				ret[index] = res
			}
			wg.Done()
		}(index, symbol)
	}

	wg.Wait()

	select {
	case err := <-errs:
		return nil, err
	default:
		return convert(ret)
	}
}

func convert(r []*Response) (*StockData, error) {
	if len(r) == 0 {
		return nil, errors.New("empty slice")
	}

	var symbols []Symbol
	for _, respo := range r {
		resp := respo
		if resp == nil || len(resp.Chart.Result) == 0 {
			continue
		}

		sym := Symbol{
			Name:   resp.Chart.Result[0].Meta.Symbol,
			Link:   fmt.Sprintf("%s/%s", linkBasePath, resp.Chart.Result[0].Meta.Symbol),
			Values: pkg.DeepCopyReverseSlice(resp.Chart.Result[0].Indicators.AdjClose[0].AdjClose),
		}
		symbols = append(symbols, sym)
	}

	return &StockData{
		Timestamps: pkg.DeepCopyReverseSlice(r[0].Chart.Result[0].Timestamp),
		Symbols:    symbols,
	}, nil
}
