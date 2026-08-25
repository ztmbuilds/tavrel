package api

import (
	"net/http"
	"noname/internal/config"
	"noname/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewAPPHttpServer(store store.Store, cfg *config.Config) *http.Server {

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Route("/api/v1/", func(r chi.Router) {
		r.Get("/traces", Make(newListTracesHandler(store)))
		r.Get("/traces/{id}", Make(newGetTraceHandler(store)))

		r.Get("/logs", Make(newListLogsHandler(store)))
		r.Get("/spans", Make(newListSpansHandler(store)))
	})

	return &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}
}
