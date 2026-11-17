package api

import (
	"github.com/ztmbuilds/crimp/internal/config"
	"github.com/ztmbuilds/crimp/internal/store"
	"github.com/ztmbuilds/crimp/internal/web"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewAPPHttpServer(store store.Store, cfg *config.Config) *http.Server {

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	authHandler := newAuthHandler(cfg)

	r.Route("/api/v1/", func(r chi.Router) {
		r.Use(authHandler.middleware)

		r.Get("/auth/session", Make(authHandler.session))
		r.Post("/auth/login", Make(authHandler.login))
		r.Post("/auth/logout", Make(authHandler.logout))

		r.Get("/traces", Make(newListTracesHandler(store)))
		r.Get("/traces/{id}", Make(newGetTraceHandler(store)))

		r.Get("/logs", Make(newListLogsHandler(store)))
		r.Get("/spans", Make(newListSpansHandler(store)))
		r.Get("/spans/links", Make(newListSpanLinksHandler(store)))
		r.Get("/services", Make(newListServicesHandler(store)))
		r.Get("/service-graph", Make(newServiceGraphHandler(store)))
		r.Get("/stats", Make(newStatsHandler(store)))
		r.Get("/environments", Make(newListEnvironmentsHandler(store)))
	})

	r.Handle("/*", web.Handler())

	return &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}
}
