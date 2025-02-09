package serve

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.HandlerFunc, level int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		if level < -2 || level > 9 {
			slog.Warn("using DefaultCompression level due to invalid level supplied", "level", level)
			level = gzip.DefaultCompression
		}
		gz, _ := gzip.NewWriterLevel(w, level)
		defer func() {
			_ = gz.Close()
		}()

		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzr, r)
	}
}
