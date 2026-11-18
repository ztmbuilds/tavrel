package otlphttp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ztmbuilds/tavrel/internal/config"
	"github.com/ztmbuilds/tavrel/internal/otlp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func failureLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		if ww.Status() < http.StatusBadRequest {
			return
		}

		slog.Warn("OTLP/HTTP request failed",
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

func authMiddleware(token string) func(http.Handler) http.Handler {
	if token == "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !otlp.TokensEqual(token, r.Header.Get("X-API-Key")) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func decompressRequestBody(maxBytes int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body io.Reader = r.Body

			switch enc := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Encoding"))); enc {
			case "", "identity":
				// no compression
			case "gzip":
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					writeError(w, r, http.StatusBadRequest, fmt.Sprintf("failed to read gzip request body: %v", err), nil)
					return
				}
				body = gz
			default:
				writeError(w, r, http.StatusUnsupportedMediaType, "unsupported Content-Encoding: "+enc, nil)
				return
			}

			data, err := io.ReadAll(io.LimitReader(body, int64(maxBytes)+1))
			if err != nil {
				writeError(w, r, http.StatusBadRequest, fmt.Sprintf("failed to read request body: %v", err), nil)
				return
			}

			if len(data) > maxBytes {
				writeError(w, r, http.StatusRequestEntityTooLarge, "request body exceeds limit", nil)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(data))
			r.ContentLength = int64(len(data))

			next.ServeHTTP(w, r)
		})
	}
}

func NewOTLPHttpServer(worker worker, traceProcessor traceExportProcessor, logsProcessor logExportProcessor, cfg *config.Config) *http.Server {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(failureLoggingMiddleware)
	r.Use(authMiddleware(cfg.IngestToken))
	r.Use(middleware.Compress(5, "application/x-protobuf", "application/json"))
	r.Use(decompressRequestBody(cfg.MaxRequestBodyBytes))

	r.Route("/v1", func(r chi.Router) {

		r.Post("/traces", newIngestTraceExportHandler(worker, traceProcessor))
		r.Post("/logs", newIngestLogsExportHandler(worker, logsProcessor))
	})

	return &http.Server{
		Addr:    ":" + cfg.OTLPHTTPPort,
		Handler: r,
	}

}
