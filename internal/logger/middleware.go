package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type respWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *respWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *respWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

func HTTPLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &respWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			path := chi.RouteContext(r.Context()).RoutePattern()
			if path == "" {
				path = r.URL.Path
			}
			status := rw.status
			if status == 0 {
				status = http.StatusOK
			}

			lvl := slog.LevelInfo
			switch {
			case status >= 500:
				lvl = slog.LevelError
			case status >= 400:
				lvl = slog.LevelWarn
			}

			log.LogAttrs(r.Context(), lvl, "http_request",
				slog.String("method", r.Method),
				slog.String("path", path),
				slog.Int("status", status),
				slog.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000.0),
				slog.String("request_id", middleware.GetReqID(r.Context())),
				slog.String("remote", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.Int("bytes", rw.bytes),
			)
		})
	}
}
