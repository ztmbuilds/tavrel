package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreLargeBatches(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

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
