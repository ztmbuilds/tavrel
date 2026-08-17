package api

import (
	"context"
	"noname/internal/store"
)

type spanStore interface {
	ListSpans(ctx context.Context, options store.ListSpansOptions) (store.Page[store.Span, store.ListSpansCursor], error)
}

func newListSpansHandler(s spanStore) AppHandler {

	return listHandler(s.ListSpans, "Spans retrieved successfully")
}
