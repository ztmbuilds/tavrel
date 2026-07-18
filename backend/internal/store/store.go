package store

import (
	"context"
	"time"
)

type Store interface {
	Close() error
	StoreTrace(ctx context.Context, trace Trace, spans []Span) error
	StoreLogs(ctx context.Context, logs []Log) error
}

type Trace struct {
	TraceID     string
	RootService string
	Environment string
	RootName    string
	StartTime   time.Time
	EndTime     time.Time
	SpanCount   int64
	Error       bool
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

type SeverityNumber int32

const (
	SeverityNumberUnspecified SeverityNumber = iota
	SeverityNumberTrace
	SeverityNumberTrace2
	SeverityNumberTrace3
	SeverityNumberTrace4
	SeverityNumberDebug
	SeverityNumberDebug2
	SeverityNumberDebug3
	SeverityNumberDebug4
	SeverityNumberInfo
	SeverityNumberInfo2
	SeverityNumberInfo3
	SeverityNumberInfo4
	SeverityNumberWarn
	SeverityNumberWarn2
	SeverityNumberWarn3
	SeverityNumberWarn4
	SeverityNumberError
	SeverityNumberError2
	SeverityNumberError3
	SeverityNumberError4
	SeverityNumberFatal
	SeverityNumberFatal2
	SeverityNumberFatal3
	SeverityNumberFatal4
)

type Log struct {
	ID                 int64
	TraceID            string
	SpanID             string
	Timestamp          time.Time
	SeverityText       string
	SeverityNumber     SeverityNumber
	Body               string
	Service            string
	Environment        string
	Attributes         map[string]any
	ResourceAttributes map[string]any
}
