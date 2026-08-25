package api

import (
	"context"
	"net/http"
	"noname/internal/store"
)

type spanStore interface {
	ListSpans(ctx context.Context, options store.ListSpansOptions) (store.Page[store.Span, store.ListSpansCursor], error)
}

func newListSpansHandler(s spanStore) AppHandler {

	return listHandler(newSpanListOptions, s.ListSpans, "Spans retrieved successfully")
}

func newSpanListOptions(r *http.Request) store.ListSpansOptions {
	options := store.ListSpansOptions{ListOptions: newListOptions[store.ListSpansCursor](r)}
	if v := r.URL.Query().Get("traceId"); v != "" {
		options.TraceID = v
	}
	return options
}
