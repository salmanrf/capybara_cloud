package middleware

import (
	"log/slog"
	"net/http"
)

type res_spy struct {
	http.ResponseWriter
	status int
}

func (w *res_spy) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func CreateLoggingMiddleware(logger *slog.Logger) func (http.Handler) http.Handler {
	logging_middleware := func (next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res := &res_spy{ResponseWriter: w}

			next.ServeHTTP(res, r)

			logger.Info("Served request", "method", r.Method, "path", r.URL.Path, "status", res.status)
		})
	}

	return logging_middleware
}