package store

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ztmbuilds/tavrel/internal/jsonvalue"
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

func mustResolveResource(t *testing.T, s *SQLiteStore, attributes map[string]any) Resource {
	t.Helper()
	resource, err := s.ResolveResource(context.Background(), mustAttributes(t, attributes))
	if err != nil {
		t.Fatal(err)
	}
	return resource
}

func mustAttributes(t *testing.T, values map[string]any) jsonvalue.Attributes {
	t.Helper()

	raw, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}

	var attributes jsonvalue.Attributes
	if err := json.Unmarshal(raw, &attributes); err != nil {
		t.Fatal(err)
	}
	return attributes
}

func assertJSON(t *testing.T, got, want any) {
	t.Helper()

	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("got %s, want %s", gotJSON, wantJSON)
	}
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
		resource := mustResolveResource(t, s, map[string]any{"service.name": service})
		if e := s.StoreTrace(ctx, tr, []Span{{
			SpanID:     "span-" + id,
			TraceID:    id,
			Service:    service,
			Name:       name,
			StartTime:  ts,
			EndTime:    ts,
			StatusCode: status,
			ResourceID: resource.ID,
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

	resourceA := mustResolveResource(t, s, map[string]any{"service.name": "svc-a"})
	resourceB := mustResolveResource(t, s, map[string]any{"service.name": "svc-b"})
	if err := s.StoreLogs(ctx, []Log{
		{TraceID: "t1", SpanID: "span-t1", Timestamp: base, SeverityNumber: 9, Body: jsonvalue.String("hello"), Service: "svc-a", ResourceID: resourceA.ID},
		{TraceID: "t2", SpanID: "span-t2", Timestamp: base, SeverityNumber: 17, Body: jsonvalue.String("boom"), Service: "svc-b", ResourceID: resourceB.ID},
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
	parentResource := mustResolveResource(t, s, map[string]any{"service.name": "parent-service"})
	childResource := mustResolveResource(t, s, map[string]any{"service.name": "child-service"})

	spans := []Span{
		{
			SpanID:     "parent-span",
			TraceID:    trace.TraceID,
			Service:    "parent-service",
			StartTime:  parentStart,
			EndTime:    parentEnd,
			ResourceID: parentResource.ID,
		},
		{
			SpanID:       "child-span",
			TraceID:      trace.TraceID,
			ParentSpanID: "parent-span",
			Service:      "child-service",
			StartTime:    childStart,
			EndTime:      childStart.Add(time.Second),
			ResourceID:   childResource.ID,
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
	resource := mustResolveResource(t, s, map[string]any{"service.name": "svc"})

	spans := make([]Span, 0, 3000)
	for i := range 3000 {
		spans = append(spans, Span{
			SpanID:     fmt.Sprintf("span-%d", i),
			TraceID:    trace.TraceID,
			Service:    "svc",
			Name:       "op",
			StartTime:  now,
			EndTime:    now,
			ResourceID: resource.ID,
		})
	}
	if err := s.StoreTrace(ctx, trace, spans); err != nil {
		t.Fatal(err)
	}
	_, gotSpans, _, err := s.GetTrace(ctx, trace.TraceID)
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
			Body:           jsonvalue.String("msg"),
			Service:        "svc",
			ResourceID:     resource.ID,
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

func TestHashAttributesDeterministic(t *testing.T) {
	firstHash, firstRaw, err := hashAttributes(mustAttributes(t, map[string]any{
		"service.name": "svc",
		"nested":       map[string]any{"b": true, "a": "x"},
		"list":         []any{map[string]any{"z": 1.0, "y": "two"}},
	}))
	if err != nil {
		t.Fatal(err)
	}

	secondHash, secondRaw, err := hashAttributes(mustAttributes(t, map[string]any{
		"list":         []any{map[string]any{"y": "two", "z": 1.0}},
		"nested":       map[string]any{"a": "x", "b": true},
		"service.name": "svc",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if firstHash != secondHash || string(firstRaw) != string(secondRaw) {
		t.Fatalf("hash is not deterministic: %q != %q", firstHash, secondHash)
	}
}

func TestResolveResourceConcurrent(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	attributes := mustAttributes(t, map[string]any{"service.name": "svc", "host.name": "host-1"})

	type result struct {
		id  int64
		err error
	}

	const goroutines = 16
	results := make(chan result, goroutines)
	var wg sync.WaitGroup

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resource, err := s.ResolveResource(ctx, attributes)
			results <- result{id: resource.ID, err: err}
		}()
	}

	wg.Wait()
	close(results)

	var id int64
	for res := range results {
		if res.err != nil {
			t.Fatal(res.err)
		}
		if id == 0 {
			id = res.id
		} else if res.id != id {
			t.Fatalf("got resource ids %d and %d, want one id", id, res.id)
		}
	}
	if id == 0 {
		t.Fatal("no resource id returned")
	}

	var count int
	if err := s.readDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM resources`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d resources, want 1", count)
	}
}

func TestStoreLogRequiresResource(t *testing.T) {
	s := openTestStore(t)

	err := s.StoreLogs(context.Background(), []Log{{
		Timestamp:      time.Now(),
		SeverityNumber: 9,
		Body:           jsonvalue.String("hello"),
	}})
	if err == nil {
		t.Fatal("expected foreign key error for missing resource")
	}
}

func TestStoreSpanRequiresResource(t *testing.T) {
	s := openTestStore(t)
	now := time.Now()

	err := s.StoreTrace(context.Background(), Trace{
		TraceID:   "trace-fk",
		StartTime: now,
		EndTime:   now,
		SpanCount: 1,
	}, []Span{{
		SpanID:    "span-fk",
		TraceID:   "trace-fk",
		Name:      "op",
		StartTime: now,
		EndTime:   now,
	}})
	if err == nil {
		t.Fatal("expected foreign key error for missing resource")
	}
}

func TestResolveResourceStoresTextAttributes(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	resource := mustResolveResource(t, s, map[string]any{"service.name": "svc"})

	var columnType string
	if err := s.readDB.QueryRowContext(ctx, `SELECT typeof(attributes) FROM resources WHERE id = ?`, resource.ID).Scan(&columnType); err != nil {
		t.Fatal(err)
	}
	if columnType != "text" {
		t.Fatalf("attributes column type = %q, want text", columnType)
	}

	var serviceName string
	if err := s.readDB.QueryRowContext(ctx, `SELECT json_extract(attributes, '$."service.name"') FROM resources WHERE id = ?`, resource.ID).Scan(&serviceName); err != nil {
		t.Fatal(err)
	}
	if serviceName != "svc" {
		t.Fatalf("json_extract(service.name) = %q, want svc", serviceName)
	}
}

func TestResolveResourceNilAttributes(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	resource, err := s.ResolveResource(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resource.Attributes != nil {
		t.Fatalf("attributes = %+v, want nil", resource.Attributes)
	}

	raw, err := json.Marshal(resource.Attributes)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `null` {
		t.Fatalf("attributes = %s, want null", raw)
	}

	var stored string
	if err := s.readDB.QueryRowContext(ctx, `SELECT attributes FROM resources WHERE id = ?`, resource.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != "null" {
		t.Fatalf("stored attributes = %q, want null", stored)
	}
}

func TestLoadResourcesBatches(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	ids := make([]int64, 0, resourceLookupBatchSize+1)
	for i := range resourceLookupBatchSize + 1 {
		resource := mustResolveResource(t, s, map[string]any{"index": i})
		ids = append(ids, resource.ID)
	}

	resources, err := s.loadResources(ctx, nil, ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != len(ids) {
		t.Fatalf("got %d resources, want %d", len(resources), len(ids))
	}
}

func TestJSONAttributesRoundTrip(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	now := time.Now()
	resource := mustResolveResource(t, s, map[string]any{"service.name": "svc"})

	trace := Trace{
		TraceID:     "trace-json",
		RootService: "svc",
		RootName:    "root",
		StartTime:   now,
		EndTime:     now,
		SpanCount:   1,
	}

	spanAttributes := mustAttributes(t, map[string]any{
		"string": "x",
		"int":    int64(1),
		"float":  1.5,
		"bool":   true,
		"null":   nil,
		"array":  []any{int64(1), "x"},
		"object": map[string]any{"nested": int64(2)},
	})
	eventAttributes := mustAttributes(t, map[string]any{"event": "value"})
	linkAttributes := mustAttributes(t, map[string]any{"link": "value"})

	spans := []Span{{
		SpanID:     "span-json",
		TraceID:    trace.TraceID,
		Service:    "svc",
		Name:       "op",
		StartTime:  now,
		EndTime:    now,
		Attributes: spanAttributes,
		Events: []Event{{
			Time:       now,
			Name:       "event",
			Attributes: eventAttributes,
		}},
		Links: []Link{{
			TraceID:    "linked-trace",
			SpanID:     "linked-span",
			Attributes: linkAttributes,
		}},
		ResourceID: resource.ID,
	}}

	if err := s.StoreTrace(ctx, trace, spans); err != nil {
		t.Fatal(err)
	}

	_, gotSpans, _, err := s.GetTrace(ctx, trace.TraceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotSpans) != 1 {
		t.Fatalf("got %d spans, want 1", len(gotSpans))
	}
	assertJSON(t, gotSpans[0].Attributes, spanAttributes)
	if len(gotSpans[0].Events) != 1 {
		t.Fatalf("got %d events, want 1", len(gotSpans[0].Events))
	}
	assertJSON(t, gotSpans[0].Events[0].Attributes, eventAttributes)

	links, err := s.ListSpanLinks(ctx, trace.TraceID, "span-json")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}
	assertJSON(t, links[0].Attributes, linkAttributes)

	logAttributes := mustAttributes(t, map[string]any{"log": "value"})
	if err := s.StoreLogs(ctx, []Log{{
		Timestamp:      now,
		SeverityNumber: 9,
		Body:           jsonvalue.String("hello"),
		Service:        "svc",
		Attributes:     logAttributes,
		ResourceID:     resource.ID,
	}}); err != nil {
		t.Fatal(err)
	}

	page, err := s.ListLogs(ctx, ListLogsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("got %d logs, want 1", len(page.Items))
	}
	body, ok := page.Items[0].Body.AsString()
	if !ok || body != "hello" {
		t.Fatalf("log body = %q, %v, want hello, true", body, ok)
	}
	assertJSON(t, page.Items[0].Attributes, logAttributes)
}

func TestListLogsHydratesResourceAttributes(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	resource := mustResolveResource(t, s, map[string]any{
		"service.name": "svc",
		"host.name":    "pod-1",
	})

	if err := s.StoreLogs(ctx, []Log{{
		Timestamp:      time.Now(),
		SeverityNumber: 9,
		Body:           jsonvalue.String("hello"),
		Service:        "svc",
		ResourceID:     resource.ID,
	}}); err != nil {
		t.Fatal(err)
	}

	page, err := s.ListLogs(ctx, ListLogsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("got %d logs, want 1", len(page.Items))
	}
	host, ok := page.Items[0].ResourceAttributes["host.name"].AsString()
	if !ok || host != "pod-1" {
		t.Fatalf("resource attributes: %+v", page.Items[0].ResourceAttributes)
	}
}

func TestDashboardErrorLogsHydrateResourceAttributes(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	now := time.Now()
	resource := mustResolveResource(t, s, map[string]any{
		"service.name": "svc",
		"host.name":    "pod-1",
	})

	if err := s.StoreLogs(ctx, []Log{{
		Timestamp:      now,
		SeverityNumber: 17,
		Body:           jsonvalue.String("boom"),
		Service:        "svc",
		ResourceID:     resource.ID,
	}}); err != nil {
		t.Fatal(err)
	}

	stats, err := s.GetDashboardStats(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(stats.RecentErrorLogs) != 1 {
		t.Fatalf("got %d error logs, want 1", len(stats.RecentErrorLogs))
	}
	host, ok := stats.RecentErrorLogs[0].ResourceAttributes["host.name"].AsString()
	if !ok || host != "pod-1" {
		t.Fatalf("resource attributes: %+v", stats.RecentErrorLogs[0].ResourceAttributes)
	}
}

func TestGetTraceReturnsResources(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	now := time.Now()

	first := mustResolveResource(t, s, map[string]any{"service.name": "svc-a", "host.name": "pod-a"})
	second := mustResolveResource(t, s, map[string]any{"service.name": "svc-b"})

	trace := Trace{
		TraceID:     "trace-resources",
		RootService: "svc-a",
		RootName:    "root",
		StartTime:   now,
		EndTime:     now,
		SpanCount:   2,
	}
	spans := []Span{
		{
			SpanID:     "span-a",
			TraceID:    trace.TraceID,
			Service:    "svc-a",
			Name:       "a",
			StartTime:  now,
			EndTime:    now,
			ResourceID: first.ID,
		},
		{
			SpanID:     "span-b",
			TraceID:    trace.TraceID,
			Service:    "svc-b",
			Name:       "b",
			StartTime:  now,
			EndTime:    now,
			ResourceID: second.ID,
		},
	}
	if err := s.StoreTrace(ctx, trace, spans); err != nil {
		t.Fatal(err)
	}

	_, gotSpans, resources, err := s.GetTrace(ctx, trace.TraceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotSpans) != 2 {
		t.Fatalf("got %d spans, want 2", len(gotSpans))
	}
	if len(resources) != 2 {
		t.Fatalf("got %d resources, want 2", len(resources))
	}
	if resources[0].ID != first.ID || resources[1].ID != second.ID {
		t.Fatalf("got resource ids %d and %d, want %d and %d", resources[0].ID, resources[1].ID, first.ID, second.ID)
	}
	host, ok := resources[0].Attributes["host.name"].AsString()
	if !ok || host != "pod-a" {
		t.Fatalf("resource attributes: %+v", resources[0].Attributes)
	}
}
