package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func openTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestListFilters(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	base := time.Now().Add(-time.Hour)

	storeTrace := func(id, service, name string, err bool, offset time.Duration) {
		t.Helper()
		ts := base.Add(offset)
		tr := Trace{
			TraceID:     id,
			RootService: service,
			RootName:    name,
			StartTime:   ts,
			EndTime:     ts,
			SpanCount:   1,
			Error:       err,
		}
		status := StatusCodeOk
		if err {
			status = StatusCodeError
		}
		if e := s.StoreTrace(ctx, tr, []Span{{
			SpanID:     "span-" + id,
			TraceID:    id,
			Service:    service,
			Name:       name,
			StartTime:  ts,
			EndTime:    ts,
			StatusCode: status,
		}}); e != nil {
			t.Fatal(e)
		}
	}
	storeTrace("t1", "svc-a", "GET /a", false, time.Minute)
	storeTrace("t2", "svc-b", "GET /b", true, 2*time.Minute)
	storeTrace("t3", "svc-a", "POST /c", true, 3*time.Minute)

	errTrue := true
	page, err := s.ListTraces(ctx, ListTracesOptions{ErrorFilter: &errTrue})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[0].TraceID != "t3" || page.Items[1].TraceID != "t2" {
		t.Fatalf("error filter: got %+v", page.Items)
	}

	page, err = s.ListTraces(ctx, ListTracesOptions{
		ListOptions: ListOptions[ListTracesCursor]{Limit: 1},
		Services:    []string{"svc-a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].TraceID != "t3" || !page.HasMore || page.NextCursor == nil {
		t.Fatalf("service filter page 1: %+v hasMore=%v cursor=%v", page.Items, page.HasMore, page.NextCursor)
	}
	page2, err := s.ListTraces(ctx, ListTracesOptions{
		ListOptions: ListOptions[ListTracesCursor]{Limit: 1, Cursor: page.NextCursor},
		Services:    []string{"svc-a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Items) != 1 || page2.Items[0].TraceID != "t1" || page2.HasMore {
		t.Fatalf("service filter page 2: %+v hasMore=%v", page2.Items, page2.HasMore)
	}

	errStatus := StatusCodeError
	spans, err := s.ListSpans(ctx, ListSpansOptions{TraceID: "t2", Status: &errStatus})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans.Items) != 1 || spans.Items[0].SpanID != "span-t2" {
		t.Fatalf("span trace+status filter: %+v", spans.Items)
	}

	if err := s.StoreLogs(ctx, []Log{
		{TraceID: "t1", SpanID: "span-t1", Timestamp: base, SeverityNumber: 9, Body: "hello", Service: "svc-a"},
		{TraceID: "t2", SpanID: "span-t2", Timestamp: base, SeverityNumber: 17, Body: "boom", Service: "svc-b"},
	}); err != nil {
		t.Fatal(err)
	}
	level := LogLevelError
	logs, err := s.ListLogs(ctx, ListLogsOptions{Level: &level, Search: "boom"})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs.Items) != 1 || logs.Items[0].TraceID != "t2" {
		t.Fatalf("log level+search filter: %+v", logs.Items)
	}
}

func TestGetServiceGraphIncludesParentServiceWhenParentStartedBeforeWindow(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	childStart := time.Now().Truncate(time.Second)
	parentStart := childStart.Add(-time.Hour)
	parentEnd := childStart.Add(time.Minute)

	trace := Trace{
		TraceID:   "trace-graph",
		StartTime: parentStart,
		EndTime:   parentEnd,
		SpanCount: 2,
	}
	spans := []Span{
		{
			SpanID:    "parent-span",
			TraceID:   trace.TraceID,
			Service:   "parent-service",
			StartTime: parentStart,
			EndTime:   parentEnd,
		},
		{
			SpanID:       "child-span",
			TraceID:      trace.TraceID,
			ParentSpanID: "parent-span",
			Service:      "child-service",
			StartTime:    childStart,
			EndTime:      childStart.Add(time.Second),
		},
	}
	if err := s.StoreTrace(ctx, trace, spans); err != nil {
		t.Fatal(err)
	}

	start := childStart.Add(-time.Minute)
	end := childStart.Add(time.Minute)
	graph, err := s.GetServiceGraph(ctx, &start, &end)
	if err != nil {
		t.Fatal(err)
	}

	wantNodes := []ServiceNode{{ID: "child-service"}, {ID: "parent-service"}}
	if len(graph.Nodes) != len(wantNodes) {
		t.Fatalf("got nodes %+v, want %+v", graph.Nodes, wantNodes)
	}
	for i, want := range wantNodes {
		if graph.Nodes[i] != want {
			t.Fatalf("got nodes %+v, want %+v", graph.Nodes, wantNodes)
		}
	}

	wantEdges := []ServiceEdge{{From: "parent-service", To: "child-service"}}
	if len(graph.Edges) != len(wantEdges) {
		t.Fatalf("got edges %+v, want %+v", graph.Edges, wantEdges)
	}
	for i, want := range wantEdges {
		if graph.Edges[i] != want {
			t.Fatalf("got edges %+v, want %+v", graph.Edges, wantEdges)
		}
	}
}

func TestStoreLargeBatches(t *testing.T) {
	s := openTestStore(t)

	ctx := context.Background()
	now := time.Now()

	trace := Trace{
		TraceID:     "trace-large",
		RootService: "svc",
		RootName:    "op",
		StartTime:   now,
		EndTime:     now,
		SpanCount:   3000,
	}
	spans := make([]Span, 0, 3000)
	for i := range 3000 {
		spans = append(spans, Span{
			SpanID:    fmt.Sprintf("span-%d", i),
			TraceID:   trace.TraceID,
			Service:   "svc",
			Name:      "op",
			StartTime: now,
			EndTime:   now,
		})
	}
	if err := s.StoreTrace(ctx, trace, spans); err != nil {
		t.Fatal(err)
	}
	_, gotSpans, err := s.GetTrace(ctx, trace.TraceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotSpans) != len(spans) {
		t.Fatalf("got %d spans, want %d", len(gotSpans), len(spans))
	}

	logs := make([]Log, 0, 4000)
	for range 4000 {
		logs = append(logs, Log{
			TraceID:        trace.TraceID,
			Timestamp:      now,
			SeverityNumber: 9,
			Body:           "msg",
			Service:        "svc",
		})
	}
	if err := s.StoreLogs(ctx, logs); err != nil {
		t.Fatal(err)
	}
	page, err := s.ListLogs(ctx, ListLogsOptions{
		ListOptions: ListOptions[ListLogsCursor]{Limit: -1},
		TraceID:     trace.TraceID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != len(logs) {
		t.Fatalf("got %d logs, want %d", len(page.Items), len(logs))
	}
}
