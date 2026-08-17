package api

import (
	"context"
	"noname/internal/store"
)

type traceStore interface {
	ListTraces(ctx context.Context, options store.ListTracesOptions) (store.Page[store.Trace, store.ListTracesCursor], error)
}

func newListTracesHandler(ts traceStore) AppHandler {

	return listHandler(ts.ListTraces, "Traces retrieved successfully")

}
