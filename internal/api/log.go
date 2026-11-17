package api

import (
	"context"
	"net/http"
	"github.com/ztmbuilds/crimp/internal/store"
)

const defaultLogsLimit = 30

type logStore interface {
	ListLogs(ctx context.Context, options store.ListLogsOptions) (store.Page[store.Log, store.ListLogsCursor], error)
}

func newListLogsHandler(ls logStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		options, err := newLogListOptions(r)
		if err != nil {
			return err
		}

		page, err := ls.ListLogs(r.Context(), options)
		if err != nil {
			return Internal(err)
		}

		writeListResponse(w, "Logs retrieved successfully", page)

		return nil
	}
}

func newLogListOptions(r *http.Request) (store.ListLogsOptions, error) {
	options := store.ListLogsOptions{
		ListOptions: newListOptions[store.ListLogsCursor](r),
	}
	query := r.URL.Query()

	if query.Get("limit") == "" {
		options.Limit = defaultLogsLimit
	}
	if v := query.Get("search"); v != "" {
		options.Search = v
	}
	if v := query.Get("traceId"); v != "" {
		options.TraceID = v
	}
	if v := query.Get("spanId"); v != "" {
		options.SpanID = v
	}

	options.Services = queryValues(r, "service")
	options.Environments = queryValues(r, "environment")

	switch v := query.Get("level"); v {
	case "", "all":
	case "info":
		level := store.LogLevelInfo
		options.Level = &level
	case "warn":
		level := store.LogLevelWarn
		options.Level = &level
	case "error":
		level := store.LogLevelError
		options.Level = &level
	default:
		return store.ListLogsOptions{}, BadRequest("invalid level: must be all, info, warn or error", nil)
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		return store.ListLogsOptions{}, err
	}
	options.Start, options.End = start, end

	return options, nil
}
