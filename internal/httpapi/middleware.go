package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/thezmc/edmotion/internal/admin"
	"github.com/thezmc/edmotion/internal/challenge"
	"log/slog"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func defaultMiddlewares(state *admin.State, requestLimitPerMinute int) chi.Middlewares {
	return chi.Middlewares{
		middleware.Recoverer,
		middleware.RealIP,
		middleware.RequestID,
		httprate.LimitByIP(requestLimitPerMinute, 1*time.Minute),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
				next.ServeHTTP(rw, r)
				slog.Info("request completed",
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
					slog.String("remoteAddr", r.RemoteAddr),
					slog.String("requestID", middleware.GetReqID(r.Context())),
					slog.Int("status", rw.statusCode),
				)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r = r.WithContext(challenge.ContextWithFixedFiles(r.Context(), state.ServeFixedFiles()))
				next.ServeHTTP(w, r)
			})
		},
	}
}
