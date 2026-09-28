package api

import (
	"context"
	"net/http"
	"noname/internal/store"
	"strconv"
)

const defaultSpansLimit = 25

type spanStore interface {
	ListSpans(ctx context.Context, options store.ListSpansOptions) (store.Page[store.Span, store.ListSpansCursor], error)
}

func newListSpansHandler(s spanStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		options, err := newSpanListOptions(r)
		if err != nil {
			return err
		}

		page, err := s.ListSpans(r.Context(), options)
		if err != nil {
			return Internal(err)
		}

		writeListResponse(w, "Spans retrieved successfully", page)

		return nil
	}
}

func newSpanListOptions(r *http.Request) (store.ListSpansOptions, error) {
	options := store.ListSpansOptions{
		ListOptions: newListOptions[store.ListSpansCursor](r),
	}
	query := r.URL.Query()

	if query.Get("limit") == "" {
		options.Limit = defaultSpansLimit
	}
	if v := query.Get("traceId"); v != "" {
		options.TraceID = v
	}
	if v := query.Get("search"); v != "" {
		options.Search = v
	}

	options.Services = queryValues(r, "service")
	options.Environments = queryValues(r, "environment")

	switch v := query.Get("status"); v {
	case "", "all":
	case "error":
		status := store.StatusCodeError
		options.Status = &status
	case "ok":
		status := store.StatusCodeOk
		options.Status = &status
	default:
		return store.ListSpansOptions{}, BadRequest("invalid status: must be all, error or ok", nil)
	}

	if v := query.Get("kind"); v != "" && v != "all" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return store.ListSpansOptions{}, BadRequest("invalid kind: must be all or an OTLP span kind (0-5)", err)
		}
		kind := store.ToSpanKind(int32(n))
		options.Kind = &kind
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		return store.ListSpansOptions{}, err
	}
	options.Start, options.End = start, end

	return options, nil
}
