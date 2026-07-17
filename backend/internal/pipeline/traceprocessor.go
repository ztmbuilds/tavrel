package pipeline

import (
	"context"
	"encoding/hex"
	"log/slog"
	"noname/internal/store"
	"time"

	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

type traceStore interface {
	StoreTrace(ctx context.Context, trace store.Trace, spans []store.Span) error
}

type TraceExportProcessor struct {
	store traceStore
}

func NewTraceExportProcessor(store traceStore) *TraceExportProcessor {
	return &TraceExportProcessor{
		store: store,
	}
}

func (tp *TraceExportProcessor) Process(req *tracepb.ExportTraceServiceRequest) {

	type traceData struct {
		spans []store.Span
		trace store.Trace
	}

	traceGroups := make(map[string]traceData)

	for _, rs := range req.ResourceSpans {
		var serviceName string
		var environment string

		for _, attr := range rs.Resource.Attributes {
			switch attr.Key {
			case "service.name":
				serviceName = attr.Value.GetStringValue()

			case "deployment.environment.name":
				environment = attr.Value.GetStringValue()

			// Backwards compatibility with older instrumentation
			case "deployment.environment":
				if environment == "" {
					environment = attr.Value.GetStringValue()
				}
			}

			// Stop once we've found everything we care about.
			if serviceName != "" && environment != "" {
				break
			}
		}
		for _, ss := range rs.ScopeSpans {
			for _, s := range ss.Spans {

				if isInvalidSpanID(s.SpanId) || isInvalidTraceID(s.TraceId) {
					continue
				}

				spanAttributes := make(map[string]any)

				for _, attr := range s.Attributes {
					spanAttributes[attr.Key] = anyValueToGo(attr.Value)
				}

				var events []store.Event
				for _, e := range s.Events {
					eventAttributes := make(map[string]any)

					for _, attr := range e.Attributes {
						eventAttributes[attr.Key] = anyValueToGo(attr.Value)
					}

					events = append(events, store.Event{
						Time:       time.Unix(0, int64(e.TimeUnixNano)),
						Name:       e.Name,
						Attributes: eventAttributes,
					})
				}

				var statusCode store.StatusCode
				var statusMessage string

				if s.Status != nil {
					statusCode = store.StatusCode(s.Status.Code)
					statusMessage = s.Status.Message
				}
				span := store.Span{
					SpanID:        hex.EncodeToString(s.SpanId),
					TraceID:       hex.EncodeToString(s.TraceId),
					ParentSpanID:  hex.EncodeToString(s.ParentSpanId),
					Service:       serviceName,
					Name:          s.Name,
					Kind:          store.SpanKind(s.Kind),
					StartTime:     time.Unix(0, int64(s.StartTimeUnixNano)),
					EndTime:       time.Unix(0, int64(s.EndTimeUnixNano)),
					StatusCode:    statusCode,
					StatusMessage: statusMessage,
					Attributes:    spanAttributes,
					Events:        events,
				}

				td, ok := traceGroups[span.TraceID]
				if !ok {
					td = traceData{
						trace: store.Trace{
							TraceID: span.TraceID,
						},
					}

				}

				td.trace.SpanCount++
				if td.trace.Environment == "" && environment != "" {
					td.trace.Environment = environment
				}

				if td.trace.StartTime.IsZero() {
					td.trace.StartTime = span.StartTime
				} else if td.trace.StartTime.After(span.StartTime) {
					td.trace.StartTime = span.StartTime
				}

				if td.trace.EndTime.IsZero() {
					td.trace.EndTime = span.EndTime
				} else if td.trace.EndTime.Before(span.EndTime) {
					td.trace.EndTime = span.EndTime
				}

				if span.ParentSpanID == "" || isInvalidSpanID(s.ParentSpanId) {

					if td.trace.RootName == "" && span.Name != "" {
						td.trace.RootName = span.Name
					}

					if td.trace.RootService == "" && span.Service != "" {
						td.trace.RootService = span.Service
					}
				}

				if !td.trace.Error && span.StatusCode == store.StatusCodeError {
					td.trace.Error = true
				}

				td.spans = append(td.spans, span)

				traceGroups[span.TraceID] = td

			}
		}
	}

	for _, v := range traceGroups {

		err := tp.store.StoreTrace(context.Background(), v.trace, v.spans)

		if err != nil {
			slog.Error("failed to store trace", "error", err)
		}
	}

}
