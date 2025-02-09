package serve

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/config"
	"jaytaylor.com/html2text"

	"github.com/rs/zerolog/log"
)

type HttpServer struct {
	datasource Datasource
	httpConfig config.HttpConfig
}

type Datasource interface {
	GetData(ctx context.Context) (*internal.Data, error)
	Name() string
}

func NewServer(datasource Datasource, conf config.HttpConfig) (*HttpServer, error) {
	return &HttpServer{
		httpConfig: conf,
		datasource: datasource,
	}, nil
}

func (h *HttpServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	data, err := h.datasource.GetData(r.Context())
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=120") // 3600 seconds = 60 minutes
	w.Header().Set("Expires", time.Now().Add(2*time.Minute).Format(http.TimeFormat))
	_, _ = w.Write(data.RenderedDefaultTemplate)
}

func (h *HttpServer) text(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	data, err := h.datasource.GetData(r.Context())
	if err != nil {
		w.WriteHeader(500)
		return
	}

	var dataToRender []byte
	if len(data.RenderedSimplifiedTemplate) > 0 {
		dataToRender = data.RenderedSimplifiedTemplate
	} else {
		dataToRender = data.RenderedDefaultTemplate
	}
	text, err := html2text.FromString(string(dataToRender), html2text.Options{
		PrettyTables:        true,
		PrettyTablesOptions: nil,
		OmitLinks:           true,
		TextOnly:            false,
	})
	if err != nil {
		w.WriteHeader(500)
		return
	}
	_, _ = w.Write([]byte(text))
}

func (h *HttpServer) Run(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	mux := http.NewServeMux()
	if h.httpConfig.UseGzip {
		mux.HandleFunc(h.httpConfig.ServePath, makeGzipHandler(h.handler, h.httpConfig.GzipCompressionLevel))
		mux.HandleFunc("/text", makeGzipHandler(h.text, h.httpConfig.GzipCompressionLevel))
	} else {
		mux.HandleFunc(h.httpConfig.ServePath, h.handler)
		mux.HandleFunc("/text", h.text)
	}

	server := http.Server{
		Addr:              h.httpConfig.Address,
		Handler:           mux,
		ReadTimeout:       3 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       3 * time.Second,
	}

	log.Info().Msg("Starting server")

	errChan := make(chan error)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warn().Err(err).Msg("Received error")
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("Shutting down server")
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
