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
	ListSpanLinks(ctx context.Context, traceID, spanID string) ([]SpanLink, error)
	ListServices(ctx context.Context) ([]string, error)
	ListEnvironments(ctx context.Context) ([]string, error)
	GetServiceGraph(ctx context.Context, start, end *time.Time) (ServiceGraph, error)
	GetDashboardStats(ctx context.Context, start, end time.Time) (DashboardStats, error)

	GetTrace(ctx context.Context, traceID string) (*Trace, []Span, error)

	DeleteOldData(ctx context.Context, cutoff time.Time) error
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

type ListTracesOptions struct {
	ListOptions[ListTracesCursor]
	Search        string
	ErrorFilter   *bool
	Services      []string
	Environments  []string
	Start         *time.Time
	End           *time.Time
	Ascending     bool
	MinDurationMs *int64
	MaxDurationMs *int64
}
type ListLogsOptions struct {
	ListOptions[ListLogsCursor]
	Search       string
	Level        *LogLevel
	Services     []string
	Environments []string
	TraceID      string
	SpanID       string
	Start        *time.Time
	End          *time.Time
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type ListSpansOptions struct {
	ListOptions[ListSpansCursor]
	TraceID      string
	Search       string
	Status       *StatusCode
	Kind         *SpanKind
	Services     []string
	Environments []string
	Start        *time.Time
	End          *time.Time
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

func ToSpanKind(i int32) SpanKind {
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
	Environment   string         `json:"environment"`
	Name          string         `json:"name"`
	Kind          SpanKind       `json:"kind"`
	StatusCode    StatusCode     `json:"statusCode"`
	StatusMessage string         `json:"statusMessage"`
	StartTime     time.Time      `json:"startTime"`
	EndTime       time.Time      `json:"endTime"`
	Attributes    map[string]any `json:"attributes"`
	Events        []Event        `json:"events"`
	Links         []Link         `json:"links,omitempty"`
}

type Link struct {
	TraceID    string         `json:"traceId"`
	SpanID     string         `json:"spanId"`
	Attributes map[string]any `json:"attributes"`
}

type SpanLink struct {
	TraceID    string          `json:"traceId"`
	SpanID     string          `json:"spanId"`
	Attributes map[string]any  `json:"attributes"`
	Target     *SpanLinkTarget `json:"target"`
}

type SpanLinkTarget struct {
	Name       string     `json:"name"`
	Service    string     `json:"service"`
	StatusCode StatusCode `json:"statusCode"`
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

type ServiceNode struct {
	ID string `json:"id"`
}

type ServiceEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ServiceGraph struct {
	Nodes []ServiceNode `json:"nodes"`
	Edges []ServiceEdge `json:"edges"`
}

type DashboardOverview struct {
	TraceCount         int64 `json:"traceCount"`
	ErrorTraceCount    int64 `json:"errorTraceCount"`
	ActiveServiceCount int64 `json:"activeServiceCount"`
}

type TraceTimeBucket struct {
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	OkCount    int64     `json:"okCount"`
	ErrorCount int64     `json:"errorCount"`
}

type DashboardStats struct {
	Overview          DashboardOverview `json:"overview"`
	TraceBuckets      []TraceTimeBucket `json:"traceBuckets"`
	RecentErrorTraces []Trace           `json:"recentErrorTraces"`
	RecentErrorLogs   []Log             `json:"recentErrorLogs"`
}
