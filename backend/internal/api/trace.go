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

		meta := &Meta{}
		if nextCursor, ok := encodeCursor(page.NextCursor); ok {

			meta.NextCursor = &nextCursor
		}

		writeJson(w, http.StatusOK, &Response{
			Message: "Traces retrieved successfully",
			Data:    page.Items,
			Meta:    meta,
		})

		return nil

	}
}
