package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/soerenschneider/aether/internal/config"
	"github.com/soerenschneider/aether/internal/datasource"
	"github.com/soerenschneider/aether/internal/datasource/alertmanager"
	"github.com/soerenschneider/aether/internal/datasource/cached"
	"github.com/soerenschneider/aether/internal/datasource/caldav"
	"github.com/soerenschneider/aether/internal/datasource/carddav"
	"github.com/soerenschneider/aether/internal/datasource/static"
	"github.com/soerenschneider/aether/internal/datasource/stocks"
	"github.com/soerenschneider/aether/internal/datasource/taskwarrior"
	"github.com/soerenschneider/aether/internal/datasource/weather"
	"github.com/soerenschneider/aether/internal/serve"

	"github.com/go-co-op/gocron"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

type deps struct {
	// the main datasource that holds the stitched html
	mainDatasource *static.StaticDatasource

	// all configured datasources, used to render all individual parts of the html that is later stitched together
	datasources []datasource.Datasource

	email      *serve.Email
	cron       *gocron.Scheduler
	httpServer *serve.HttpServer
}

func (d *deps) Cleanup() {
	if d.cron != nil {
		d.cron.Stop()
	}
}

func (d *deps) HasEmailSupport() bool {
	return d.email != nil
}

func wrapDatasource(ds datasource.Datasource, conf config.DatasourceConfig) (datasource.Datasource, error) {
	if conf.IsCached() {
		var opts []cached.Opts
		if conf.GetCacheExpiry() > 0 {
			opts = append(opts, cached.WithRefreshInterval(conf.GetCacheExpiry()))
		}

		return cached.New(ds, opts...)
	}

	return ds, nil
}

func buildDatasources(conf config.Config) ([]datasource.Datasource, error) {
	var datasources []datasource.Datasource
	var errs error

	for _, dsConfig := range conf.Datasources {
		var err error
		var ds datasource.Datasource

		switch dsConfig.Config.Type() {
		case config.Alertmanager:
			ds, err = buildAlertmanager(dsConfig.Config.(*config.AlertmanagerConfig))
		case config.CalDav:
			ds, err = buildCalDav(dsConfig.Config.(*config.CalDavConfig))
		case config.CardDav:
			ds, err = buildCardDav(dsConfig.Config.(*config.CardDavConfig))
		case config.Stocks:
			ds, err = buildStocks(dsConfig.Config.(*config.StocksConfig))
		case config.Taskwarrior:
			ds, err = buildTaskwarrior(dsConfig.Config.(*config.TaskwarriorConfig))
		case config.Weather:
			ds, err = buildWeather(dsConfig.Config.(*config.WeatherConfig))
		default:
			return nil, fmt.Errorf("unknown datasource: %q", dsConfig.Config.Type())
		}

		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			wrapped, err := wrapDatasource(ds, dsConfig.Config)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				datasources = append(datasources, wrapped)
			}
		}
	}

	return datasources, errs
}

func buildWeather(conf *config.WeatherConfig) (*weather.WeatherDatasource, error) {
	opts := []weather.Opt{
		weather.WithHttpClient(httpClient),
	}

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, weather.WithTemplateFile(conf.TemplateFile))
	}

	weather, err := weather.New(conf.ApiKey, conf.Latitude, conf.Longitude, opts...)
	if err != nil {
		return nil, err
	}

	return weather, nil
}

func buildAlertmanager(conf *config.AlertmanagerConfig) (*alertmanager.AlertmanagerDatasource, error) {
	var opts []alertmanager.Opt
	if len(conf.BasePath) > 0 {
		opts = append(opts, alertmanager.WithBasePath(conf.BasePath))
	}

	if len(conf.Scheme) > 0 {
		opts = append(opts, alertmanager.WithScheme(conf.Scheme))
	}

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, alertmanager.WithTemplateFile(conf.TemplateFile))
	}

	return alertmanager.New(conf.Host, opts...)
}

func buildStocks(conf *config.StocksConfig) (*stocks.StocksDatasource, error) {
	opts := []stocks.Opts{
		stocks.WithHttpClient(httpClient),
	}

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, stocks.WithTemplateFile(conf.TemplateFile))
	}

	return stocks.New(conf.Symbols, opts...)
}

func buildTaskwarrior(conf *config.TaskwarriorConfig) (*taskwarrior.TaskwarriorDatasource, error) {
	var opts []taskwarrior.Opt

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, taskwarrior.WithTemplateFile(conf.TemplateFile))
	}

	return taskwarrior.New(opts...)
}

func buildCardDav(conf *config.CardDavConfig) (*carddav.CarddavDatasource, error) {
	opts := []carddav.Opt{
		carddav.WithHttpClient(httpClient),
	}

	if len(conf.Password) > 0 && len(conf.Username) > 0 {
		opts = append(opts, carddav.WithBasicAuth(conf.Username, conf.Password))
	}

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, carddav.WithTemplateFile(conf.TemplateFile))
	}

	return carddav.New(conf.Endpoint, opts...)
}

func buildCalDav(conf *config.CalDavConfig) (*caldav.CaldavDatasource, error) {
	opts := []caldav.Opt{
		caldav.WithHttpClient(httpClient),
	}

	if len(conf.Password) > 0 && len(conf.Username) > 0 {
		opts = append(opts, caldav.WithBasicAuth(conf.Username, conf.Password))
	}

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, caldav.WithTemplateFile(conf.TemplateFile))
	}

	return caldav.New(conf.Endpoint, opts...)
}

func init() {
	getHttpClient()
}

var httpClient *http.Client
var onceHttp sync.Once

func getHttpClient() *http.Client {
	onceHttp.Do(func() {
		client := retryablehttp.NewClient()
		client.RetryMax = 3
		client.RetryWaitMax = 1 * time.Second
		client.Logger = &ZerologAdapter{}
		httpClient = client.HTTPClient
	})

	return httpClient
}

func buildEmail(conf config.EmailConfig) (*serve.Email, error) {
	var opts []serve.EmailOpt
	if len(conf.Password) > 0 {
		opts = append(opts, serve.WithPassword(conf.UserName, conf.Password))
	}

	return serve.NewEmail(conf.From, conf.To, conf.Host, opts...)
}

type ZerologAdapter struct {
}

func (z *ZerologAdapter) Debug(msg string, keysAndValues ...interface{}) {
	log.Debug().Str("checker", "prometheus").Interface("details", keysAndValues).Msg(msg)
}

func (z *ZerologAdapter) Info(msg string, keysAndValues ...interface{}) {
	log.Info().Str("checker", "prometheus").Interface("details", keysAndValues).Msg(msg)
}

func (z *ZerologAdapter) Warn(msg string, keysAndValues ...interface{}) {
	log.Warn().Str("checker", "prometheus").Interface("details", keysAndValues).Msg(msg)
}

func (z *ZerologAdapter) Error(msg string, keysAndValues ...interface{}) {
	log.Error().Str("checker", "prometheus").Interface("details", keysAndValues).Msg(msg)
}
