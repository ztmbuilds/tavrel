package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"noname/internal/store"
	"slices"
	"time"
)

type traceStore interface {
	ListTraces(ctx context.Context, options store.ListTracesOptions) (store.Page[store.Trace, store.ListTracesCursor], error)
	GetTrace(ctx context.Context, traceID string) (*store.Trace, []store.Span, error)
}

type spanNode struct {
	Span     store.Span  `json:"span"`
	Children []*spanNode `json:"children"`
}
type TraceData struct {
	Trace *store.Trace `json:"trace"`
	Spans []*spanNode  `json:"spans"`
}

const defaultTracesLimit = 20

func newListTracesHandler(ts traceStore) AppHandler {

	return func(w http.ResponseWriter, r *http.Request) error {
		options, err := newTraceListOptions(r)
		if err != nil {
			return err
		}

		page, err := ts.ListTraces(r.Context(), options)
		if err != nil {
			return Internal(err)
		}

		writeListResponse(w, "Traces retrieved successfully", page)

		return nil
	}

}

func newTraceListOptions(r *http.Request) (store.ListTracesOptions, error) {
	options := store.ListTracesOptions{
		ListOptions: newListOptions[store.ListTracesCursor](r),
	}
	query := r.URL.Query()

	if query.Get("limit") == "" {
		options.Limit = defaultTracesLimit
	}
	if v := query.Get("search"); v != "" {
		options.Search = v
	}

	options.Services = queryValues(r, "service")
	options.Environments = queryValues(r, "environment")

	switch v := query.Get("status"); v {
	case "", "all":
	case "error":
		err := true
		options.ErrorFilter = &err
	case "ok":
		err := false
		options.ErrorFilter = &err
	default:
		return store.ListTracesOptions{}, BadRequest("invalid status: must be all, error or ok", nil)
	}

	switch v := query.Get("order"); v {
	case "", "desc":
	case "asc":
		options.Ascending = true
	default:
		return store.ListTracesOptions{}, BadRequest("invalid order: must be asc or desc", nil)
	}

	if v := query.Get("start"); v != "" {
		start, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return store.ListTracesOptions{}, BadRequest("invalid start: must be ISO 8601 ", err)
		}
		options.Start = &start
	}
	if v := query.Get("end"); v != "" {
		end, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return store.ListTracesOptions{}, BadRequest("invalid end: must be ISO 8601 ", err)
		}
		options.End = &end
	}
	if options.Start != nil && options.End != nil && options.Start.After(*options.End) {
		return store.ListTracesOptions{}, BadRequest("invalid time range: start must not be after end", nil)
	}

	return options, nil
}

func newGetTraceHandler(ts traceStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {

		traceId := r.PathValue("id")
		trace, spans, err := ts.GetTrace(r.Context(), traceId)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return NotFound("trace not found", err)
			}
			return Internal(err)
		}

		nodes := buildTraceTree(spans)

		writeJson(w, http.StatusOK, &Response{
			Message: "Trace retrieved successfully",
			Data: TraceData{
				Trace: trace,
				Spans: nodes,
			},
		})

		return nil
	}
}

func buildTraceTree(spans []store.Span) []*spanNode {

	if len(spans) == 0 {
		return nil
	}

	nodes := make(map[string]*spanNode, len(spans))

	for _, span := range spans {

		nodes[span.SpanID] = &spanNode{
			Span:     span,
			Children: []*spanNode{},
		}

	}

	roots := make([]*spanNode, 0, len(spans))

	for _, n := range nodes {
		if n.Span.ParentSpanID == "" {
			roots = append(roots, n)
			continue
		}

		p, ok := nodes[n.Span.ParentSpanID]

		if !ok {
			// parent span might not exist yet so just treat as root.
			roots = append(roots, n)
			continue
		}

		p.Children = append(p.Children, n)

	}

	slices.SortFunc(roots, func(a, b *spanNode) int {
		return a.Span.StartTime.Compare(b.Span.StartTime)
	})

	for _, root := range roots {
		sortSpanChildren(root)
	}

	return roots
}

func sortSpanChildren(node *spanNode) {
	slices.SortFunc(node.Children, func(a, b *spanNode) int {
		return a.Span.StartTime.Compare(b.Span.StartTime)
	})
	for _, child := range node.Children {
		sortSpanChildren(child)
	}
}
