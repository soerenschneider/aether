package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/internal/config"
	"github.com/soerenschneider/aether/internal/datasource/alertmanager"
	"github.com/soerenschneider/aether/internal/datasource/astral"
	"github.com/soerenschneider/aether/internal/datasource/cached"
	"github.com/soerenschneider/aether/internal/datasource/caldav"
	"github.com/soerenschneider/aether/internal/datasource/carddav"
	"github.com/soerenschneider/aether/internal/datasource/static"
	"github.com/soerenschneider/aether/internal/datasource/taskwarrior"
	"github.com/soerenschneider/aether/internal/datasource/weather"
	"github.com/soerenschneider/aether/internal/serve"
	"github.com/soerenschneider/aether/internal/templates"
	"go.uber.org/multierr"
)

type deps struct {
	// the main datasource that holds the stitched html
	mainDatasource *static.StaticDatasource

	// all configured datasources, used to render all individual parts of the html that is later stitched together
	datasources []Datasource

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

func wrapDatasource(ds Datasource, conf config.DatasourceConfig) (Datasource, error) {
	if conf.IsCached() {
		var opts []cached.Opts
		if conf.GetCacheExpiry() > 0 {
			opts = append(opts, cached.WithRefreshInterval(conf.GetCacheExpiry()))
		}

		return cached.New(ds, opts...)
	}

	return ds, nil
}

func buildDatasources(conf config.Config, _ *sync.WaitGroup) ([]Datasource, error) {
	var datasources []Datasource
	var errs error

	for _, dsConfig := range conf.Datasources {

		var err error
		var ds Datasource

		switch dsConfig.Config.Type() {
		case config.Alertmanager:
			ds, err = buildAlertmanager(dsConfig.Config.(*config.AlertmanagerConfig))
		case config.Astral:
			ds, err = buildAstral(dsConfig.Config.(*config.AstralConfig))
		case config.CalDav:
			ds, err = buildCalDav(dsConfig.Config.(*config.CalDavConfig))
		case config.CardDav:
			ds, err = buildCardDav(dsConfig.Config.(*config.CardDavConfig))
		//case config.Stocks:
		//	ds, err = buildStocks(dsConfig.Config.(*config.StocksConfig))
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

func buildAstral(conf *config.AstralConfig) (*astral.Astral, error) {
	var err error
	templateData := templates.TemplateData{}
	templateData.DefaultTemplate, err = templates.GetTemplate("astral/default.html")
	if err != nil {
		return nil, err
	}

	tz, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return nil, err
	}
	impl, err := astral.New(astral.Lat(conf.Latitude), astral.Lon(conf.Longitude), templateData, tz)
	if err != nil {
		return nil, err
	}

	return impl, nil
}

func buildWeather(conf *config.WeatherConfig) (*weather.WeatherDatasource, error) {
	clientOpts := []weather.OpenweatherMapOpt{
		weather.WithHttpClient(httpClient),
	}

	if conf.Count > 0 {
		clientOpts = append(clientOpts, weather.WithCount(conf.Count))
	}

	apiKey := conf.ApiKey
	if len(conf.ApiKeyFile) > 0 {
		content, err := os.ReadFile(conf.ApiKeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not api key from file %q: %w", conf.ApiKeyFile, err)
		}
		apiKey = string(content)
	}

	weatherClient, err := weather.NewOpenweatherMapClient(apiKey, weather.Lat(conf.Latitude), weather.Lon(conf.Longitude), conf.NiceName, clientOpts...)
	if err != nil {
		return nil, err
	}

	templateData := templates.TemplateData{}
	templateData.DefaultTemplate, err = templates.GetTemplate("weather/default.html")
	if err != nil {
		return nil, err
	}
	templateData.SimpleTemplate, err = templates.GetTemplate("weather/simple.html")
	if err != nil {
		return nil, err
	}

	weatherProvider, err := weather.New(weatherClient, templateData)
	if err != nil {
		return nil, err
	}

	return weatherProvider, nil
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

	var err error
	templateData := templates.TemplateData{}
	templateData.DefaultTemplate, err = templates.GetTemplate("alertmanager/default.html")
	if err != nil {
		return nil, err
	}

	if conf.SimpleTemplateFile != "" {
		templateData.SimpleTemplate, err = os.ReadFile(conf.SimpleTemplateFile)
		if err != nil {
			return nil, err
		}
	}

	return alertmanager.New(conf.Host, templateData, opts...)
}

func buildTaskwarrior(conf *config.TaskwarriorConfig) (*taskwarrior.Datasource, error) {
	var opts []taskwarrior.Opt

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, taskwarrior.WithTemplateFile(conf.TemplateFile))
	}

	if conf.SummaryDays > 0 {
		opts = append(opts, taskwarrior.WithSummaryDays(conf.SummaryDays))
	}

	if conf.Limit > 0 {
		opts = append(opts, taskwarrior.WithLimit(conf.Limit))

	}

	client, err := taskwarrior.NewTaskwarriorClient(conf.TaskRcFile)
	if err != nil {
		return nil, err
	}

	templateData := templates.TemplateData{}
	templateData.DefaultTemplate, err = templates.GetTemplate("taskwarrior/default.html")
	if err != nil {
		return nil, err
	}

	return taskwarrior.New(client, templateData, opts...)
}

func buildCardDav(conf *config.CardDavConfig) (*carddav.CarddavDatasource, error) {
	opts := []carddav.Opt{
		carddav.WithHttpClient(httpClient),
	}

	if (len(conf.Password) > 0 || len(conf.PasswordFile) > 0) && len(conf.Username) > 0 {
		password := conf.Password
		if len(conf.PasswordFile) > 0 {
			content, err := os.ReadFile(conf.PasswordFile)
			if err != nil {
				return nil, fmt.Errorf("could not read password from file %q: %w", conf.PasswordFile, err)
			}
			password = string(content)
		}
		opts = append(opts, carddav.WithBasicAuth(conf.Username, password))
	}

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, carddav.WithTemplateFile(conf.TemplateFile))
	}

	templateData := templates.TemplateData{}
	var err error
	templateData.DefaultTemplate, err = templates.GetTemplate("contacts/default.html")
	if err != nil {
		return nil, err
	}
	templateData.SimpleTemplate, err = templates.GetTemplate("contacts/simple.html")
	if err != nil {
		return nil, err
	}
	return carddav.New(conf.Endpoint, templateData, opts...)
}

func buildCalDav(conf *config.CalDavConfig) (*caldav.CaldavDatasource, error) {
	opts := []caldav.Opt{
		caldav.WithHttpClient(httpClient),
	}

	if (len(conf.Password) > 0 || len(conf.PasswordFile) > 0) && len(conf.Username) > 0 {
		password := conf.Password
		if len(conf.PasswordFile) > 0 {
			content, err := os.ReadFile(conf.PasswordFile)
			if err != nil {
				return nil, fmt.Errorf("could not read password from file %q: %w", conf.PasswordFile, err)
			}
			password = string(content)
		}
		opts = append(opts, caldav.WithBasicAuth(conf.Username, password))
	}

	if len(conf.TemplateFile) > 0 {
		opts = append(opts, caldav.WithRegularTemplateFile(conf.TemplateFile))
	}

	templateData := templates.TemplateData{}
	var err error
	templateData.DefaultTemplate, err = templates.GetTemplate("calendar/default.html")
	if err != nil {
		return nil, err
	}
	templateData.SimpleTemplate, err = templates.GetTemplate("calendar/simple.html")
	if err != nil {
		return nil, err
	}
	return caldav.New(conf.Endpoint, templateData, opts...)
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
	var errs error
	var opts []serve.EmailOpt

	username, err := conf.GetUsername()
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	password, err := conf.GetPassword()
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	from, err := conf.GetFrom()
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	recipients, err := conf.GetRecipients()
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	if errs != nil {
		return nil, errs
	}
	return serve.NewEmail(from, recipients, conf.Host, username, password, opts...)
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
