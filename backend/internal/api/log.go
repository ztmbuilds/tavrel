package api

import (
	"context"
	"noname/internal/store"
)

type logStore interface {
	ListLogs(ctx context.Context, options store.ListLogsOptions) (store.Page[store.Log, store.ListLogsCursor], error)
}

func newListLogsHandler(ls logStore) AppHandler {

	return listHandler(ls.ListLogs, "Logs retrieved successfully")
}
