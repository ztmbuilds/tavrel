package api

import (
	"context"
	"net/http"
	"github.com/ztmbuilds/tavrel/internal/store"
	"time"
)

type serviceStore interface {
	ListServices(ctx context.Context) ([]string, error)
	GetServiceGraph(ctx context.Context, start, end *time.Time) (store.ServiceGraph, error)
}

func newListServicesHandler(s serviceStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		services, err := s.ListServices(r.Context())
		if err != nil {
			return Internal(err)
		}
		if services == nil {
			services = []string{}
		}

		writeJson(w, http.StatusOK, &Response{
			Message: "Services retrieved successfully",
			Data:    services,
			Meta:    &Meta{},
		})

		return nil
	}
}

func newServiceGraphHandler(s serviceStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		start, end, err := parseTimeRange(r)
		if err != nil {
			return err
		}

		graph, err := s.GetServiceGraph(r.Context(), start, end)
		if err != nil {
			return Internal(err)
		}

		writeJson(w, http.StatusOK, &Response{
			Message: "Service graph retrieved successfully",
			Data:    graph,
			Meta:    &Meta{},
		})

		return nil
	}
}
