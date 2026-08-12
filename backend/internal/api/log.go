package api

import (
	"context"
	"net/http"
	"noname/internal/store"
)

type logStore interface {
	ListLogs(ctx context.Context, options store.ListLogsOptions) (store.Page[store.Log, store.ListLogsCursor], error)
}

func newLogHandler(ls logStore) AppHandler {

	return func(w http.ResponseWriter, r *http.Request) error {

		options := newListOptions[store.ListLogsCursor](r)

		page, err := ls.ListLogs(r.Context(), options)

		if err != nil {
			return Internal(err)
		}

		writeListResponse(w, "Logs retrieved successfully", page)

		return nil
	}
}
