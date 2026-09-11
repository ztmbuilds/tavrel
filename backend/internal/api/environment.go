package api

import (
	"context"
	"net/http"
)

type environmentStore interface {
	ListEnvironments(ctx context.Context) ([]string, error)
}

func newListEnvironmentsHandler(s environmentStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		environments, err := s.ListEnvironments(r.Context())
		if err != nil {
			return Internal(err)
		}
		if environments == nil {
			environments = []string{}
		}

		writeJson(w, http.StatusOK, &Response{
			Message: "Environments retrieved successfully",
			Data:    environments,
			Meta:    &Meta{},
		})

		return nil
	}
}
