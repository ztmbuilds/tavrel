package api

import (
	"context"
	"net/http"
	"noname/internal/store"
)

type traceStore interface {
	ListTraces(ctx context.Context, options store.ListTracesOptions) (store.Page[store.Trace, store.ListTracesCursor], error)
}

func newListTracesHandler(ts traceStore) AppHandler {

	return func(w http.ResponseWriter, r *http.Request) error {

		options := newListOptions[store.ListTracesCursor](r)

		page, err := ts.ListTraces(r.Context(), options)

		if err != nil {
			return Internal(err)
		}

		writeListResponse(w, "Traces retrieved successfully", page)

		return nil

	}
}
