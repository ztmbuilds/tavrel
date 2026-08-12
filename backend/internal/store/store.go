package store

import (
	"context"
	"time"
)

type Store interface {
	Close() error
	StoreTrace(ctx context.Context, trace Trace, spans []Span) error
	StoreLogs(ctx context.Context, logs []Log) error
	ListTraces(ctx context.Context, options ListTracesOptions) (Page[Trace, ListTracesCursor], error)
	ListLogs(ctx context.Context, options ListLogsOptions) (Page[Log, ListLogsCursor], error)
}

type ListTracesCursor struct {
	StartTime time.Time
	TraceID   string
}

type ListLogsCursor struct {
	TimeStamp time.Time
	ID        int64
}

type ListTracesOptions = ListOptions[ListTracesCursor]
type ListLogsOptions = ListOptions[ListLogsCursor]
type Trace struct {
	TraceID     string
	RootService string
	Environment string
	RootName    string
	StartTime   time.Time
	EndTime     time.Time
	SpanCount   int64
	Error       bool
	UpdatedAt   time.Time
}

type SpanKind int32

const (
	SpanKindUnspecified SpanKind = iota
	SpanKindInternal
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

type StatusCode int32

const (
	StatusCodeUnset StatusCode = iota
	StatusCodeOk
	StatusCodeError
)

type Span struct {
	SpanID        string
	TraceID       string
	ParentSpanID  string
	Service       string
	Name          string
	Kind          SpanKind
	StatusCode    StatusCode
	StatusMessage string
	StartTime     time.Time
	EndTime       time.Time
	Attributes    map[string]any
	Events        []Event
}

type Event struct {
	Time       time.Time
	Name       string
	Attributes map[string]any
}

type Log struct {
	ID                 int64
	TraceID            string
	SpanID             string
	Timestamp          time.Time
	SeverityText       string
	SeverityNumber     int32
	Body               any
	Service            string
	Environment        string
	Attributes         map[string]any
	ResourceAttributes map[string]any
}
