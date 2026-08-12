package api

import (
	"net/http"
	"noname/internal/config"
	"noname/internal/store"

	"github.com/go-chi/chi/v5"
)

func NewAPPHttpServer(store store.Store, cfg *config.Config) *http.Server {

	r := chi.NewRouter()

	r.Route("/api/v1/", func(r chi.Router) {
		r.Get("/traces", Make(newListTracesHandler(store)))
		r.Get("/logs", Make(newLogHandler(store)))
	})

	return &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}
}
