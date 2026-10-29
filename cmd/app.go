package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/config"
	"github.com/soerenschneider/aether/internal/datasource/cached"
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
	// shell for the pages that datasources offer next to the dashboard
	pageTemplate *template.Template
	// datasources that offer a page of their own, by datasource name
	pages map[string]internal.PageProvider
}

type pageTemplateInput struct {
	Title   string
	Content template.HTML
}

type summaryFragment struct {
	Data []string
	Name string
}

type aetherTemplateInput struct {
	Summary template.HTML
	// compact summaries of the datasources, replace the summary if set
	Tiles     template.HTML
	Data      template.HTML
	UpdatedAt time.Time
	// names of the datasources that could not be fetched
	Failed []string
	// datasources that could not be updated and show older data
	Outdated []outdatedSource
}

type outdatedSource struct {
	Name string
	// when the data was fetched, e.g. "10:05"
	Since string
}

type dataPieces struct {
	RegularHtmlPieces [][]byte
	SimpleHtmlPieces  [][]byte
	SummaryPieces     []summaryFragment
	TilePieces        [][]byte
	// per datasource the time its outdated data was fetched, zero if the data is current
	OutdatedSince []time.Time
	Names         []string
	Failed        []string
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

	// the dashboard and the pages share their styles
	styles, err := templates.GetTemplate("main/styles.html")
	if err != nil {
		return nil, err
	}

	aetherTempl, err := template.New("aether").Parse(string(templateData.DefaultTemplate))
	if err != nil {
		return nil, err
	}
	if _, err := aetherTempl.Parse(string(styles)); err != nil {
		return nil, err
	}

	pageData, err := templates.GetTemplate("main/page.html")
	if err != nil {
		return nil, err
	}
	pageTempl, err := template.New("page").Parse(string(pageData))
	if err != nil {
		return nil, err
	}
	if _, err := pageTempl.Parse(string(styles)); err != nil {
		return nil, err
	}

	pages := map[string]internal.PageProvider{}
	for _, ds := range deps.datasources {
		if provider := findPageProvider(ds); provider != nil {
			pages[ds.Name()] = provider
		}
	}

	summaryTempl, err := template.New("summary").Funcs(template.FuncMap{
		"nameToId": pkg.NameToId,
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
		pageTemplate:    pageTempl,
		pages:           pages,
	}, nil
}

// findPageProvider returns the datasource, or the datasource it wraps, if it offers a page of its own.
func findPageProvider(ds any) internal.PageProvider {
	for ds != nil {
		if provider, ok := ds.(internal.PageProvider); ok && provider.PageName() != "" {
			return provider
		}

		wrapper, ok := ds.(interface{ Unwrap() cached.Datasource })
		if !ok {
			return nil
		}
		ds = wrapper.Unwrap()
	}
	return nil
}

// renderPage renders a page that a datasource offers inside the page shell.
func (a *App) renderPage(r *http.Request, provider internal.PageProvider) ([]byte, error) {
	page, err := provider.RenderPage(r)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := a.pageTemplate.Execute(&buf, pageTemplateInput{
		Title:   page.Title,
		Content: template.HTML(page.Content),
	}); err != nil {
		return nil, err
	}
	return a.Minify(buf.Bytes()), nil
}

func (a *App) pageHandler(provider internal.PageProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, err := a.renderPage(r, provider)
		if err != nil {
			log.Error().Err(err).Str("page", provider.PageName()).Msg("could not render page")
			http.Error(w, "could not render page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(page)
	}
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

	for _, provider := range a.pages {
		a.deps.httpServer.AddPage(provider.PageName(), a.pageHandler(provider))
	}

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
	tilePieces := make([][]byte, len(a.deps.datasources))
	names := make([]string, len(a.deps.datasources))
	failed := make([]bool, len(a.deps.datasources))
	outdatedSince := make([]time.Time, len(a.deps.datasources))

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)

	p := pool.New().WithErrors().WithContext(ctx).WithMaxGoroutines(8)

	defer cancel()
	start := time.Now()
	for index, ds := range a.deps.datasources {
		names[index] = ds.Name()
		f := func(ctx context.Context) error {
			start := time.Now()
			data, err := ds.GetData(ctx)
			if err != nil {
				failed[index] = true
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
			tilePieces[index] = data.RenderedTile
			if data.Outdated {
				outdatedSince[index] = data.FetchedAt
			}

			log.Debug().Msgf("Finished datasource %d (%s) after %v", index, ds.Name(), time.Since(start))
			return nil
		}
		p.Go(f)
	}

	err := p.Wait()
	log.Debug().Msgf("Updated %d datasources in %v", len(a.deps.datasources), time.Since(start))
	if err != nil {
		log.Error().Err(err).Msg("could not render all templates")
	}

	var failedNames []string
	for index, hasFailed := range failed {
		if hasFailed {
			failedNames = append(failedNames, names[index])
		}
	}

	return dataPieces{
		RegularHtmlPieces: regularHtmlPieces,
		SimpleHtmlPieces:  simplifiedHtmlPieces,
		SummaryPieces:     summaryPieces,
		TilePieces:        tilePieces,
		Names:             names,
		Failed:            failedNames,
		OutdatedSince:     outdatedSince,
	}, nil
}

// stitchTiles links every non-empty tile to the card of its datasource
func (a *App) stitchTiles(tiles [][]byte, names []string, outdatedSince []time.Time) []byte {
	htmlData := bytes.NewBuffer(nil)
	for i := 0; i < len(tiles); i++ {
		if len(bytes.TrimSpace(tiles[i])) == 0 {
			continue
		}
		// link to the page of the datasource if it has one, otherwise to its card
		href := "#" + pkg.NameToId(names[i])
		if provider, ok := a.pages[names[i]]; ok {
			href = provider.PageName()
		}
		if !outdatedSince[i].IsZero() {
			_, _ = fmt.Fprintf(htmlData, `<a class="tile outdated" href="%s">`, template.HTMLEscapeString(href))
			_, _ = htmlData.Write(tiles[i])
			_, _ = fmt.Fprintf(htmlData, `<div class="tile-outdated">Out of date · %s</div>`, formatSince(outdatedSince[i], time.Now()))
			_, _ = htmlData.WriteString("</a>")
			continue
		}
		_, _ = fmt.Fprintf(htmlData, `<a class="tile" href="%s">`, template.HTMLEscapeString(href))
		_, _ = htmlData.Write(tiles[i])
		_, _ = htmlData.WriteString("</a>")
	}

	return htmlData.Bytes()
}

// stitchPieces wraps every non-empty piece in its own card, using the datasource name as anchor
func (a *App) stitchPieces(pieces [][]byte, names []string, outdatedSince []time.Time) ([]byte, error) {
	htmlData := bytes.NewBuffer(nil)
	for i := 0; i < len(pieces); i++ {
		if len(bytes.TrimSpace(pieces[i])) == 0 {
			continue
		}
		id := template.HTMLEscapeString(pkg.NameToId(names[i]))
		_, _ = fmt.Fprintf(htmlData, `<section class="card" id="%s">`, id)
		if !outdatedSince[i].IsZero() {
			_, _ = fmt.Fprintf(htmlData, `<p class="outdated-note">Out of date · last updated %s</p>`, formatSince(outdatedSince[i], time.Now()))
		}
		_, _ = htmlData.Write(pieces[i])
		_, _ = htmlData.WriteString("</section>")
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

	updatedAt := time.Now()

	var outdated []outdatedSource
	for i, since := range data.OutdatedSince {
		if !since.IsZero() {
			outdated = append(outdated, outdatedSource{Name: data.Names[i], Since: formatSince(since, updatedAt)})
		}
	}

	regularHtmlData, err := a.stitchPieces(data.RegularHtmlPieces, data.Names, data.OutdatedSince)
	if err != nil {
		return nil, err
	}

	regularHtmlDoc := bytes.NewBuffer(nil)
	if err := a.aetherTemplate.Execute(regularHtmlDoc, aetherTemplateInput{
		Summary:   template.HTML(summaryHtmlData.String()),
		Tiles:     template.HTML(a.stitchTiles(data.TilePieces, data.Names, data.OutdatedSince)),
		Data:      template.HTML(regularHtmlData),
		UpdatedAt: updatedAt,
		Failed:    data.Failed,
		Outdated:  outdated,
	}); err != nil {
		return nil, err
	}

	simplifiedHtmlData, err := a.stitchPieces(data.SimpleHtmlPieces, data.Names, data.OutdatedSince)
	if err != nil {
		return nil, err
	}

	simpleHtmlDoc := bytes.NewBuffer(nil)
	if err := a.aetherTemplate.Execute(simpleHtmlDoc, aetherTemplateInput{
		Summary:   template.HTML(summaryHtmlData.String()),
		Data:      template.HTML(simplifiedHtmlData),
		UpdatedAt: updatedAt,
		Failed:    data.Failed,
		Outdated:  outdated,
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

// formatSince formats the time data was fetched at, with the weekday if it wasn't today.
func formatSince(t, now time.Time) string {
	if pkg.IsToday(t, now) {
		return t.Format("15:04")
	}
	return t.Format("Mon 15:04")
}
