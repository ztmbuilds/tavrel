package api

import (
	"context"
	"net/http"
	"noname/internal/store"
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
		start, end, err := parseServiceGraphWindow(r)
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

func parseServiceGraphWindow(r *http.Request) (start, end *time.Time, err error) {
	query := r.URL.Query()

	if v := query.Get("start"); v != "" {
		parsed, parseErr := time.Parse(time.RFC3339, v)
		if parseErr != nil {
			return nil, nil, BadRequest("invalid start: must be ISO 8601 ", parseErr)
		}
		start = &parsed
	}
	if v := query.Get("end"); v != "" {
		parsed, parseErr := time.Parse(time.RFC3339, v)
		if parseErr != nil {
			return nil, nil, BadRequest("invalid end: must be ISO 8601 ", parseErr)
		}
		end = &parsed
	}
	if start != nil && end != nil && start.After(*end) {
		return nil, nil, BadRequest("invalid time range: start must not be after end", nil)
	}

	return start, end, nil
}
