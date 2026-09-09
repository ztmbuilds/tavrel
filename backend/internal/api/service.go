package api

import (
	"context"
	"net/http"
)

type serviceStore interface {
	ListServices(ctx context.Context) ([]string, error)
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
