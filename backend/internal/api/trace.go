package api

import (
	"context"
	"net/http"
	"noname/internal/store"
	"strconv"
)

type traceStore interface {
	ListTraces(ctx context.Context, options store.ListTracesOptions) (store.Page[store.Trace, store.ListTracesCursor], error)
}

func newListTracesHandler(ts traceStore) AppHandler {

	return func(w http.ResponseWriter, r *http.Request) error {

		limit := 50

		if value := r.URL.Query().Get("limit"); value != "" {
			if num, err := strconv.Atoi(value); err == nil && num > 0 {

				limit = num
			}
		}

		options := store.ListTracesOptions{
			Limit:  limit,
			Cursor: nil,
		}

		encodedCursor := r.URL.Query().Get("cursor")

		cursor, ok := decodeCursor[store.ListTracesCursor](encodedCursor)

		if ok {
			options.Cursor = &cursor
		}

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
