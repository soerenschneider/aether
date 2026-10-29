//go:build dev

package main

// Development server that renders the dashboard with fake data, so the html can be worked on without any of the real
// backends. Templates are read from disk on every request: edit, save, reload the browser. Run it with `make dev`.

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/config"
	"github.com/soerenschneider/aether/internal/datasource/cached"
	"github.com/soerenschneider/aether/internal/serve"
	"github.com/soerenschneider/aether/internal/static"
	"github.com/soerenschneider/aether/internal/templates"
	"jaytaylor.com/html2text"
)

func init() {
	devMain = runDevServer
}

func runDevServer() {
	addr := flag.String("addr", "127.0.0.1:8080", "address to listen on")
	templateDir := flag.String("templates", "internal/templates", "directory to read the templates from, empty to use the embedded ones")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *templateDir != "" {
		templates.UseDir(*templateDir)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", static.Handler("/static/"))
	mux.HandleFunc("/text", devTextHandler)
	mux.HandleFunc("/calendar", devSubPageHandler)
	mux.HandleFunc("/email", devEmailHandler)
	mux.HandleFunc("/", devPageHandler)

	log.Info().Msgf("Serving fake data at http://%s, pick a scenario with ?scenario=%s", *addr, strings.Join(scenarioNames(), "|"))
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal().Err(err).Msg("dev server failed")
	}
}

func devPageHandler(w http.ResponseWriter, r *http.Request) {
	s, ok := parseScenario(w, r)
	if !ok {
		return
	}

	// simulate a server that went away after the page was loaded: htmx refreshes fail
	if s == scenarioStale && r.Header.Get("HX-Request") == "true" {
		http.Error(w, "simulated outage", http.StatusServiceUnavailable)
		return
	}

	data, problems, err := renderDev(r.Context(), s)
	if err != nil {
		writeDevError(w, err)
		return
	}

	page := data.RenderedDefaultTemplate
	if len(problems) > 0 {
		page = injectBeforeBodyEnd(page, problemsOverlay(problems))
	}
	if s == scenarioStale {
		page = injectBeforeBodyEnd(page, `<script>document.body.classList.add("stale")</script>`)
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(page)
}

// devSubPageHandler serves the pages that datasources offer next to the dashboard, e.g. the month view.
func devSubPageHandler(w http.ResponseWriter, r *http.Request) {
	s, ok := parseScenario(w, r)
	if !ok {
		return
	}

	// the page shows the data of the last update
	app, _, err := buildDevApp(r.Context(), s, true)
	if err != nil {
		writeDevError(w, err)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/")
	for _, provider := range app.pages {
		if provider.PageName() != name {
			continue
		}

		page, err := app.renderPage(r, provider)
		if err != nil {
			writeDevError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(page)
		return
	}

	http.NotFound(w, r)
}

// devEmailHandler shows the dashboard as it is sent by email. With ?inline-only the <style> element that remains after
// inlining is removed as well, as many email clients do.
func devEmailHandler(w http.ResponseWriter, r *http.Request) {
	s, ok := parseScenario(w, r)
	if !ok {
		return
	}

	data, _, err := renderDev(r.Context(), s)
	if err != nil {
		writeDevError(w, err)
		return
	}

	css, err := templates.GetTemplate("main/email.css")
	if err != nil {
		writeDevError(w, err)
		return
	}

	email, err := serve.PrepareEmail(data.RenderedDefaultTemplate, string(css), "http://"+r.Host+"/")
	if err != nil {
		writeDevError(w, err)
		return
	}
	if _, ok := r.URL.Query()["inline-only"]; ok {
		email = regexp.MustCompile(`(?s)<style.*?</style>`).ReplaceAllString(email, "")
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	_, _ = w.Write([]byte(email))
}

func devTextHandler(w http.ResponseWriter, r *http.Request) {
	s, ok := parseScenario(w, r)
	if !ok {
		return
	}

	data, _, err := renderDev(r.Context(), s)
	if err != nil {
		writeDevError(w, err)
		return
	}

	text, err := html2text.FromString(string(data.RenderedSimplifiedTemplate), html2text.Options{
		PrettyTables: true,
		OmitLinks:    true,
	})
	if err != nil {
		writeDevError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	_, _ = w.Write([]byte(text))
}

const scenarioCookie = "aether-dev-scenario"

// parseScenario reads the scenario from the query, falling back to the one chosen last. The links of the pages don't
// know about scenarios, so it's remembered in a cookie.
func parseScenario(w http.ResponseWriter, r *http.Request) (scenario, bool) {
	name := r.URL.Query().Get("scenario")
	if name == "" {
		if cookie, err := r.Cookie(scenarioCookie); err == nil {
			name = cookie.Value
		}
	}
	if name == "" {
		return scenarioNormal, true
	}

	if _, ok := scenarios[scenario(name)]; !ok {
		http.Error(w, fmt.Sprintf("unknown scenario %q, available: %s", name, strings.Join(scenarioNames(), ", ")), http.StatusBadRequest)
		return "", false
	}

	http.SetCookie(w, &http.Cookie{Name: scenarioCookie, Value: name, Path: "/", SameSite: http.SameSiteLaxMode})
	return scenario(name), true
}

// renderDev builds the datasources from scratch, so changed templates are picked up on every request. Errors of the
// individual datasources are returned as problems, as the app would otherwise just hide the failing datasource.
func renderDev(ctx context.Context, s scenario) (*internal.Data, []string, error) {
	app, recorder, err := buildDevApp(ctx, s, false)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	data, err := app.getRenderedData(ctx)
	if err != nil {
		return nil, nil, err
	}

	return data, recorder.get(), nil
}

// buildDevApp builds the app with fake datasources, optionally fetching them once so their pages have data.
func buildDevApp(ctx context.Context, s scenario, fetch bool) (*App, *errorRecorder, error) {
	recorder := &errorRecorder{}
	datasources, err := buildFakeDatasources(s, recorder)
	if err != nil {
		return nil, nil, err
	}

	mainTemplate, err := templates.GetTemplate("main/main.html")
	if err != nil {
		return nil, nil, err
	}
	summaryTemplate, err := templates.GetTemplate("main/summary.html")
	if err != nil {
		return nil, nil, err
	}

	conf := config.DefaultConfig()
	// keep the html readable for the browser's dev tools
	conf.Http.Minify = false

	app, err := NewApp(deps{datasources: datasources}, templates.TemplateData{
		DefaultTemplate: mainTemplate,
		SimpleTemplate:  summaryTemplate,
	}, &conf)
	if err != nil {
		return nil, nil, err
	}

	if fetch {
		for _, ds := range datasources {
			_, _ = ds.GetData(ctx)
		}
	}

	return app, recorder, nil
}

// errorRecorder collects unexpected errors of datasources, simulated outages are not reported.
type errorRecorder struct {
	mutex  sync.Mutex
	errors []string
}

func (e *errorRecorder) wrap(ds Datasource) Datasource {
	return &recordingDatasource{Datasource: ds, recorder: e}
}

func (e *errorRecorder) add(name string, err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.errors = append(e.errors, fmt.Sprintf("%s: %v", name, err))
}

func (e *errorRecorder) get() []string {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	sort.Strings(e.errors)
	return e.errors
}

type recordingDatasource struct {
	Datasource
	recorder *errorRecorder
}

// Unwrap makes the pages of the wrapped datasource available.
func (r *recordingDatasource) Unwrap() cached.Datasource {
	return r.Datasource
}

func (r *recordingDatasource) GetData(ctx context.Context) (*internal.Data, error) {
	data, err := r.Datasource.GetData(ctx)
	if err != nil && !errors.Is(err, errSimulatedOutage) {
		r.recorder.add(r.Name(), err)
	}
	return data, err
}

func problemsOverlay(problems []string) string {
	var b strings.Builder
	b.WriteString(`<pre style="position:fixed;left:16px;right:16px;bottom:16px;z-index:100;margin:0;padding:14px 16px;` +
		`background:#2b0d0d;color:#ffb4b4;border:1px solid #a3231f;border-radius:10px;font:13px/1.5 ui-monospace,Menlo,monospace;` +
		`white-space:pre-wrap;max-height:40vh;overflow:auto">`)
	b.WriteString("<strong>Datasource errors (dev server only)</strong>\n")
	for _, problem := range problems {
		b.WriteString(html.EscapeString(problem))
		b.WriteString("\n")
	}
	b.WriteString("</pre>")
	return b.String()
}

func injectBeforeBodyEnd(page []byte, snippet string) []byte {
	idx := bytes.LastIndex(page, []byte("</body>"))
	if idx < 0 {
		return append(page, snippet...)
	}

	ret := make([]byte, 0, len(page)+len(snippet))
	ret = append(ret, page[:idx]...)
	ret = append(ret, snippet...)
	return append(ret, page[idx:]...)
}

func writeDevError(w http.ResponseWriter, err error) {
	log.Error().Err(err).Msg("could not render page")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html><title>render error</title><body style="font:14px/1.5 ui-monospace,Menlo,monospace;padding:24px">`+
		`<h1 style="font-size:16px">Could not render the page</h1><pre style="white-space:pre-wrap">%s</pre></body>`, html.EscapeString(err.Error()))
}
