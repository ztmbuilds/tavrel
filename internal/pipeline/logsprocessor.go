package pipeline

import (
	"context"
	"encoding/hex"
	"log/slog"
	"noname/internal/jsonvalue"
	"noname/internal/store"
	"time"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type logsStore interface {
	StoreLogs(ctx context.Context, logs []store.Log) error
	ResolveResource(ctx context.Context, attributes jsonvalue.Attributes) (store.Resource, error)
}

type LogsExportProcessor struct {
	store logsStore
}

func NewLogsExportProcessor(store logsStore) *LogsExportProcessor {
	return &LogsExportProcessor{
		store: store,
	}

}

func (lp *LogsExportProcessor) Process(req *logspb.ExportLogsServiceRequest) {

	var logs []store.Log

	for _, rl := range req.ResourceLogs {
		resource, err := lp.store.ResolveResource(context.Background(), attributesToJSON(rl.GetResource().GetAttributes()))
		if err != nil {
			slog.Error("failed to resolve resource", "error", err)
			continue
		}

		service, environment := resourceIdentity(resource.Attributes)

		for _, sl := range rl.ScopeLogs {

			for _, l := range sl.LogRecords {

				var traceID []byte
				var spanID []byte

				if !isInvalidTraceID(l.TraceId) {
					traceID = l.TraceId
				}

				if !isInvalidSpanID(l.SpanId) {
					spanID = l.SpanId
				}

				logAttributes := make(jsonvalue.Attributes)

				for _, attr := range l.Attributes {
					logAttributes[attr.Key] = anyValueToJSON(attr.Value)
				}

				timeStamp := l.TimeUnixNano

				if timeStamp == 0 {
					timeStamp = l.ObservedTimeUnixNano
				}

				log := store.Log{
					SpanID:         hex.EncodeToString(spanID),
					TraceID:        hex.EncodeToString(traceID),
					Service:        service,
					Environment:    environment,
					Attributes:     logAttributes,
					SeverityText:   l.SeverityText,
					SeverityNumber: int32(l.SeverityNumber),
					Body:           anyValueToJSON(l.Body),
					Timestamp:      time.Unix(0, int64(timeStamp)),
					ResourceID:     resource.ID,
				}

				logs = append(logs, log)
			}

		}

	}

	err := lp.store.StoreLogs(context.Background(), logs)

	if err != nil {
		slog.Error("failed to store logs", "error", err)
	}

}
