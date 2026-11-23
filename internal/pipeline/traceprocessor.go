package pipeline

import (
	"context"
	"encoding/hex"
	"github.com/ztmbuilds/tavrel/internal/jsonvalue"
	"github.com/ztmbuilds/tavrel/internal/store"
	"log/slog"
	"time"

	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

type traceStore interface {
	StoreTraces(ctx context.Context, traces []store.TraceSpans) error
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

	traceGroups := make(map[string]store.TraceSpans)

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
					td = store.TraceSpans{
						Trace: store.Trace{
							TraceID: span.TraceID,
						},
					}

				}

				td.Trace.SpanCount++
				if td.Trace.Environment == "" && environment != "" {
					td.Trace.Environment = environment
				}

				if td.Trace.StartTime.IsZero() {
					td.Trace.StartTime = span.StartTime
				} else if td.Trace.StartTime.After(span.StartTime) {
					td.Trace.StartTime = span.StartTime
				}

				if td.Trace.EndTime.IsZero() {
					td.Trace.EndTime = span.EndTime
				} else if td.Trace.EndTime.Before(span.EndTime) {
					td.Trace.EndTime = span.EndTime
				}

				if span.ParentSpanID == "" || isInvalidSpanID(s.ParentSpanId) {

					if td.Trace.RootName == "" && span.Name != "" {
						td.Trace.RootName = span.Name
					}

					if td.Trace.RootService == "" && span.Service != "" {
						td.Trace.RootService = span.Service
					}
				}

				if !td.Trace.Error && span.StatusCode == store.StatusCodeError {
					td.Trace.Error = true
				}

				td.Spans = append(td.Spans, span)

				traceGroups[span.TraceID] = td

			}
		}
	}

	traces := make([]store.TraceSpans, 0, len(traceGroups))
	for _, trace := range traceGroups {
		traces = append(traces, trace)
	}

	if err := tp.store.StoreTraces(context.Background(), traces); err != nil {
		slog.Error("failed to store traces", "error", err)
	}

}
