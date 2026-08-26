package otlphttp

import (
	"log/slog"
	"net/http"
	"noname/internal/config"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("OTLP/HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration", time.Since(start).String(),
			"request_id", middleware.GetReqID(r.Context()),
			"content_type", r.Header.Get("Content-Type"),
		)
	})
}

func NewOTLPHttpServer(worker worker, traceProcessor traceExportProcessor, logsProcessor logExportProcessor, cfg *config.Config) *http.Server {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(loggingMiddleware)

	r.Route("/v1", func(r chi.Router) {

		r.Post("/traces", newIngestTraceExportHandler(worker, traceProcessor))
		r.Post("/logs", newIngestLogsExportHandler(worker, logsProcessor))
	})

	return &http.Server{
		Addr:    ":" + cfg.OTLPHTTPPort,
		Handler: r,
	}

}
