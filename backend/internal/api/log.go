package api

import (
	"context"
	"net/http"
	"noname/internal/store"
	"time"
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

	if v := query.Get("start"); v != "" {
		start, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return store.ListLogsOptions{}, BadRequest("invalid start: must be ISO 8601 ", err)
		}
		options.Start = &start
	}
	if v := query.Get("end"); v != "" {
		end, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return store.ListLogsOptions{}, BadRequest("invalid end: must be ISO 8601 ", err)
		}
		options.End = &end
	}
	if options.Start != nil && options.End != nil && options.Start.After(*options.End) {
		return store.ListLogsOptions{}, BadRequest("invalid time range: start must not be after end", nil)
	}

	return options, nil
}
