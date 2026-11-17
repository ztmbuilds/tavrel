package pipeline

import (
	"context"
	"encoding/hex"
	"log/slog"
	"github.com/ztmbuilds/crimp/internal/jsonvalue"
	"github.com/ztmbuilds/crimp/internal/store"
	"time"

	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

type traceStore interface {
	StoreTrace(ctx context.Context, trace store.Trace, spans []store.Span) error
	ResolveResource(ctx context.Context, attributes jsonvalue.Attributes) (store.Resource, error)
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
		resource, err := tp.store.ResolveResource(context.Background(), attributesToJSON(rs.GetResource().GetAttributes()))
		if err != nil {
			slog.Error("failed to resolve resource", "error", err)
			continue
		}

		serviceName, environment := resourceIdentity(resource.Attributes)

		for _, ss := range rs.ScopeSpans {
			for _, s := range ss.Spans {

				if isInvalidSpanID(s.SpanId) || isInvalidTraceID(s.TraceId) {
					continue
				}

				spanAttributes := attributesToJSON(s.Attributes)

				var events []store.Event
				for _, e := range s.Events {
					events = append(events, store.Event{
						Time:       time.Unix(0, int64(e.TimeUnixNano)),
						Name:       e.Name,
						Attributes: attributesToJSON(e.Attributes),
					})
				}

				var links []store.Link
				for _, l := range s.Links {
					if isInvalidTraceID(l.TraceId) || isInvalidSpanID(l.SpanId) {
						continue
					}

					links = append(links, store.Link{
						TraceID:    hex.EncodeToString(l.TraceId),
						SpanID:     hex.EncodeToString(l.SpanId),
						Attributes: attributesToJSON(l.Attributes),
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
					Environment:   environment,
					Name:          s.Name,
					Kind:          store.SpanKind(s.Kind),
					StartTime:     time.Unix(0, int64(s.StartTimeUnixNano)),
					EndTime:       time.Unix(0, int64(s.EndTimeUnixNano)),
					StatusCode:    statusCode,
					StatusMessage: statusMessage,
					Attributes:    spanAttributes,
					Events:        events,
					Links:         links,
					ResourceID:    resource.ID,
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
