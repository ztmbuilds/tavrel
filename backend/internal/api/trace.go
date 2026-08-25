package api

import (
	"context"
	"net/http"
	"noname/internal/store"
	"slices"
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

func newListTracesHandler(ts traceStore) AppHandler {

	return listHandler(newListOptions[store.ListTracesCursor], ts.ListTraces, "Traces retrieved successfully")

}

func newGetTraceHandler(ts traceStore) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {

		traceId := r.PathValue("id")
		trace, spans, err := ts.GetTrace(r.Context(), traceId)

		if err != nil {
			return err
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

	return roots
}
