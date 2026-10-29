package static

import (
	"embed"
	"net/http"
)

var (
	//go:embed htmx.min.js
	assets embed.FS
)

// Handler serves the embedded static assets. They only change with a new release, so browsers may cache them for a day.
func Handler(prefix string) http.Handler {
	fileServer := http.StripPrefix(prefix, http.FileServer(http.FS(assets)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		fileServer.ServeHTTP(w, r)
	})
}
