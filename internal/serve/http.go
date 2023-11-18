package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/soerenschneider/aether/internal/config"
	"github.com/soerenschneider/aether/internal/datasource"

	"github.com/rs/zerolog/log"
)

type HttpServer struct {
	datasource datasource.Datasource
	httpConfig config.HttpConfig
}

func NewServer(datasource datasource.Datasource, conf config.HttpConfig) (*HttpServer, error) {
	return &HttpServer{
		httpConfig: conf,
		datasource: datasource,
	}, nil
}

func (h *HttpServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	html, err := h.datasource.GetHtml(context.Background())
	if err != nil {
		w.WriteHeader(500)
		return
	}

	fmt.Fprint(w, html)
}

func (h *HttpServer) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(h.httpConfig.ServePath, h.handler)

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
		shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
	return nil
}
