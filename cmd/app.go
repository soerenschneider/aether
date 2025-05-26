package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/config"
	"github.com/soerenschneider/aether/internal/datasource/static"
	"github.com/soerenschneider/aether/internal/serve"
	"github.com/soerenschneider/aether/internal/templates"
	"github.com/soerenschneider/aether/pkg"
	"github.com/sourcegraph/conc/pool"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
	"jaytaylor.com/html2text"
)

type Datasource interface {
	GetData(ctx context.Context) (*internal.Data, error)
	Name() string
}

type App struct {
	deps            deps
	conf            *config.Config
	minifier        *minify.M
	aetherTemplate  *template.Template
	summaryTemplate *template.Template
}

type summaryFragment struct {
	Data []string
	Name string
}

type aetherTemplateInput struct {
	Summary template.HTML
	Data    template.HTML
}

type dataPieces struct {
	RegularHtmlPieces [][]byte
	SimpleHtmlPieces  [][]byte
	SummaryPieces     []summaryFragment
}

func NewApp(deps deps, templateData templates.TemplateData, conf *config.Config) (*App, error) {
	var minifier *minify.M
	if conf.Http.Minify {
		minifier = minify.New()
		minifier.AddFunc("text/html", html.Minify)
		minifier.AddFunc("text/css", css.Minify)
		minifier.AddFunc("image/svg+xml", svg.Minify)
		minifier.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	}

	aetherTempl, err := template.New("aether").Parse(string(templateData.DefaultTemplate))
	if err != nil {
		return nil, err
	}

	summaryTempl, err := template.New("summary").Funcs(template.FuncMap{
		"nameToId": pkg.NameToId,
		"add": func(i, j int) int {
			return i + j
		},
		"isEven": func(i int) bool {
			return i%2 == 0
		},
	}).Parse(string(templateData.SimpleTemplate))
	if err != nil {
		return nil, err
	}

	return &App{
		deps:            deps,
		conf:            conf,
		minifier:        minifier,
		aetherTemplate:  aetherTempl,
		summaryTemplate: summaryTempl,
	}, nil
}

func (a *App) Start(ctx context.Context, wg *sync.WaitGroup) error {
	if len(a.deps.datasources) == 0 {
		dieOnError(errors.New("no datasource configured"), "could not build datasources")
	}

	if a.deps.HasEmailSupport() && len(a.conf.Email.At) > 0 {
		if err := a.scheduleEmail(*a.conf); err != nil {
			return fmt.Errorf("scheduling email dispatch failed: %w", err)
		}
	}

	a.update(ctx)
	go func() {
		a.runHttpServer(ctx, *a.conf.Http, wg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)

		for {
			select {
			case <-ticker.C:
				a.update(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (a *App) runHttpServer(ctx context.Context, conf config.HttpConfig, wg *sync.WaitGroup) {
	var err error
	a.deps.httpServer, err = serve.NewServer(a.deps.mainDatasource, conf)
	dieOnError(err, "could not setup http server")

	err = a.deps.httpServer.Run(ctx, wg)
	dieOnError(err, "could not start http server")
}

func (a *App) update(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	data, err := a.getRenderedData(ctx)
	if err != nil {
		log.Error().Err(err).Msg("errors while producing html")
	}

	once.Do(func() {
		a.deps.mainDatasource = static.NewStatic(data)
	})
	a.deps.mainDatasource.Update(data)
}

func getConfig() (*config.Config, error) {
	return config.ReadConfig(flags.ConfigFile)
}

func (a *App) fetchData(ctx context.Context) (dataPieces, error) {
	regularHtmlPieces := make([][]byte, len(a.deps.datasources))
	simplifiedHtmlPieces := make([][]byte, len(a.deps.datasources))
	summaryPieces := make([]summaryFragment, len(a.deps.datasources))

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)

	p := pool.New().WithErrors().WithContext(ctx).WithMaxGoroutines(8)

	defer cancel()
	start := time.Now()
	for index, ds := range a.deps.datasources {
		f := func(ctx context.Context) error {
			start := time.Now()
			data, err := ds.GetData(ctx)
			if err != nil {
				return err
			}
			regularHtmlPieces[index] = data.RenderedDefaultTemplate
			if len(data.RenderedSimplifiedTemplate) > 0 {
				simplifiedHtmlPieces[index] = data.RenderedSimplifiedTemplate
			} else {
				simplifiedHtmlPieces[index] = data.RenderedDefaultTemplate
			}
			summaryPieces[index] = summaryFragment{
				Data: data.Summary,
				Name: ds.Name(),
			}

			log.Debug().Msgf("Finished datasource %d (%s) after %v", index, ds.Name(), time.Since(start))
			return nil
		}
		p.Go(f)
	}

	log.Debug().Msgf("Updated %d datasources in %v", len(a.deps.datasources), time.Since(start))
	if err := p.Wait(); err != nil {
		log.Error().Err(err).Msg("could not render all templates")
		return dataPieces{
			RegularHtmlPieces: regularHtmlPieces,
			SimpleHtmlPieces:  simplifiedHtmlPieces,
			SummaryPieces:     summaryPieces,
		}, nil
	}

	return dataPieces{
		RegularHtmlPieces: regularHtmlPieces,
		SimpleHtmlPieces:  simplifiedHtmlPieces,
		SummaryPieces:     summaryPieces,
	}, nil
}

func (a *App) stitchPieces(pieces [][]byte) ([]byte, error) {
	htmlData := bytes.NewBuffer(nil)
	for i := 0; i < len(pieces); i++ {
		_, _ = htmlData.Write(pieces[i])
	}

	return htmlData.Bytes(), nil
}

func (a *App) getRenderedData(ctx context.Context) (*internal.Data, error) {
	data, err := a.fetchData(ctx)
	if err != nil {
		return nil, err
	}

	summaryHtmlData := bytes.NewBuffer(nil)
	if err := a.summaryTemplate.Execute(summaryHtmlData, data.SummaryPieces); err != nil {
		return nil, err
	}

	regularHtmlData, err := a.stitchPieces(data.RegularHtmlPieces)
	if err != nil {
		return nil, err
	}

	regularHtmlDoc := bytes.NewBuffer(nil)
	if err := a.aetherTemplate.Execute(regularHtmlDoc, aetherTemplateInput{
		Summary: template.HTML(summaryHtmlData.String()),
		Data:    template.HTML(regularHtmlData),
	}); err != nil {
		return nil, err
	}

	simplifiedHtmlData, err := a.stitchPieces(data.SimpleHtmlPieces)
	if err != nil {
		return nil, err
	}

	simpleHtmlDoc := bytes.NewBuffer(nil)
	if err := a.aetherTemplate.Execute(simpleHtmlDoc, aetherTemplateInput{
		Summary: template.HTML(summaryHtmlData.String()),
		Data:    template.HTML(simplifiedHtmlData),
	}); err != nil {
		return nil, err
	}

	return &internal.Data{
		RenderedDefaultTemplate:    a.Minify(regularHtmlDoc.Bytes()),
		RenderedSimplifiedTemplate: a.Minify(simpleHtmlDoc.Bytes()),
	}, nil
}

func (a *App) Minify(input []byte) []byte {
	if a.minifier == nil {
		return input
	}

	res, _ := a.minifier.Bytes("text/html", input)
	return res
}

func (a *App) scheduleEmail(conf config.Config) error {
	location := time.UTC
	if conf.Email.Timezone != "" {
		var err error
		location, err = time.LoadLocation(conf.Email.Timezone)
		if err != nil {
			return err
		}
	}

	a.deps.cron = gocron.NewScheduler(location)
	i, err := a.deps.cron.Every(1).Day().At(conf.Email.At).Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		a.sendEmail(ctx)
	})
	if err != nil {
		return err
	}

	a.deps.cron.StartAsync()
	log.Info().Msgf("Scheduling daily email at %s, next run at %v", conf.Email.At, i.NextRun())

	return err
}

func (a *App) sendEmail(ctx context.Context) {
	data, _ := a.deps.mainDatasource.GetData(ctx)

	var dataToRender []byte
	if len(data.RenderedSimplifiedTemplate) > 0 {
		dataToRender = data.RenderedSimplifiedTemplate
	} else {
		dataToRender = data.RenderedDefaultTemplate
	}
	text, _ := html2text.FromString(string(dataToRender), html2text.Options{
		PrettyTables:        true,
		PrettyTablesOptions: nil,
		OmitLinks:           true,
		TextOnly:            false,
	})

	if err := a.deps.email.SendReport(ctx, "Aether", string(data.RenderedDefaultTemplate), text); err != nil {
		log.Error().Err(err).Msg("could not send email")
		return
	}
	log.Info().Msg("Successfully dispatched email")
}
