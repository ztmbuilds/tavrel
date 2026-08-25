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
	ListSpans(ctx context.Context, options ListSpansOptions) (Page[Span, ListSpansCursor], error)

	GetTrace(ctx context.Context, traceID string) (*Trace, []Span, error)
}

type ListTracesCursor struct {
	StartTime time.Time
	TraceID   string
}

type ListLogsCursor struct {
	TimeStamp time.Time
	ID        int64
}

type ListSpansCursor struct {
	StartTime time.Time
	SpanID    string
}

type ListTracesOptions = ListOptions[ListTracesCursor]
type ListLogsOptions = ListOptions[ListLogsCursor]
type ListSpansOptions struct {
	ListOptions[ListSpansCursor]
	TraceID string
}

type Trace struct {
	TraceID     string    `json:"traceId"`
	RootService string    `json:"rootService"`
	Environment string    `json:"environment"`
	RootName    string    `json:"rootName"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	SpanCount   int64     `json:"spanCount"`
	Error       bool      `json:"error"`
	UpdatedAt   time.Time `json:"updatedAt"`
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

func toSpanKind(i int32) SpanKind {
	switch SpanKind(i) {
	case SpanKindInternal:
		return SpanKindInternal
	case SpanKindServer:
		return SpanKindServer
	case SpanKindClient:
		return SpanKindClient
	case SpanKindProducer:
		return SpanKindProducer
	case SpanKindConsumer:
		return SpanKindConsumer
	default:
		return SpanKindUnspecified
	}
}

type StatusCode int32

const (
	StatusCodeUnset StatusCode = iota
	StatusCodeOk
	StatusCodeError
)

func toStatusCode(i int32) StatusCode {
	switch StatusCode(i) {
	case StatusCodeOk:
		return StatusCodeOk
	case StatusCodeError:
		return StatusCodeError
	default:
		return StatusCodeUnset
	}
}

type Span struct {
	SpanID        string         `json:"spanId"`
	TraceID       string         `json:"traceId"`
	ParentSpanID  string         `json:"parentSpanId"`
	Service       string         `json:"service"`
	Name          string         `json:"name"`
	Kind          SpanKind       `json:"kind"`
	StatusCode    StatusCode     `json:"statusCode"`
	StatusMessage string         `json:"statusMessage"`
	StartTime     time.Time      `json:"startTime"`
	EndTime       time.Time      `json:"endTime"`
	Attributes    map[string]any `json:"attributes"`
	Events        []Event        `json:"events"`
}

type Event struct {
	Time       time.Time      `json:"time"`
	Name       string         `json:"name"`
	Attributes map[string]any `json:"attributes"`
}

type Log struct {
	ID                 int64          `json:"id"`
	TraceID            string         `json:"traceId"`
	SpanID             string         `json:"spanId"`
	Timestamp          time.Time      `json:"timestamp"`
	SeverityText       string         `json:"severityText"`
	SeverityNumber     int32          `json:"severityNumber"`
	Body               any            `json:"body"`
	Service            string         `json:"service"`
	Environment        string         `json:"environment"`
	Attributes         map[string]any `json:"attributes"`
	ResourceAttributes map[string]any `json:"resourceAttributes"`
}
