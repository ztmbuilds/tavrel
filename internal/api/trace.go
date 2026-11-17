package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"github.com/ztmbuilds/crimp/internal/store"
	"slices"
	"strconv"
)

type traceStore interface {
	ListTraces(ctx context.Context, options store.ListTracesOptions) (store.Page[store.Trace, store.ListTracesCursor], error)
	GetTrace(ctx context.Context, traceID string) (*store.Trace, []store.Span, []store.Resource, error)
}

type spanNode struct {
	Span     store.Span  `json:"span"`
	Children []*spanNode `json:"children"`
}
type TraceData struct {
	Trace     *store.Trace     `json:"trace"`
	Spans     []*spanNode      `json:"spans"`
	Resources []store.Resource `json:"resources"`
}

const defaultTracesLimit = 20

// maxDurationFilterMs is the largest millisecond value that still fits in an
// int64 nanosecond span; larger values overflow the store's ms→ns conversion.
const maxDurationFilterMs = math.MaxInt64 / 1_000_000

func parseDurationFilterMs(query url.Values, name string) (*int64, error) {
	value := query.Get(name)
	if value == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 || int64(n) > maxDurationFilterMs {
		return nil, BadRequest(fmt.Sprintf("invalid %s: must be between 0 and %d", name, maxDurationFilterMs), err)
	}
	ms := int64(n)
	return &ms, nil
}

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

	minDurationMs, err := parseDurationFilterMs(query, "minDurationMs")
	if err != nil {
		return store.ListTracesOptions{}, err
	}
	options.MinDurationMs = minDurationMs
	maxDurationMs, err := parseDurationFilterMs(query, "maxDurationMs")
	if err != nil {
		return store.ListTracesOptions{}, err
	}
	options.MaxDurationMs = maxDurationMs
	if options.MinDurationMs != nil && options.MaxDurationMs != nil && *options.MinDurationMs > *options.MaxDurationMs {
		return store.ListTracesOptions{}, BadRequest("invalid duration range: minDurationMs must not be greater than maxDurationMs", nil)
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		return store.ListTracesOptions{}, err
	}
	options.Start, options.End = start, end

	return options, nil
}

func newGetTraceHandler(ts traceStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {

		traceId := r.PathValue("id")
		trace, spans, resources, err := ts.GetTrace(r.Context(), traceId)

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
				Trace:     trace,
				Spans:     nodes,
				Resources: resources,
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
