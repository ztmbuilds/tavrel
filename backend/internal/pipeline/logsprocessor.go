package pipeline

import (
	"context"
	"encoding/hex"
	"log/slog"
	"noname/internal/store"
	"time"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type logsStore interface {
	StoreLogs(ctx context.Context, logs []store.Log) error
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
		var service string
		var environment string

		resourceAttributes := make(map[string]any)

		for _, attr := range rl.Resource.Attributes {
			resourceAttributes[attr.Key] = anyValueToGo(attr.Value)

			switch attr.Key {
			case "service.name":
				if service == "" {
					service = attr.Value.GetStringValue()
				}

			case "deployment.environment.name":
				if environment == "" {
					environment = attr.Value.GetStringValue()
				}

			// Backwards compatibility
			case "deployment.environment":
				if environment == "" {
					environment = attr.Value.GetStringValue()
				}
			}
		}

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

				logAttributes := make(map[string]any)

				for _, attr := range l.Attributes {
					logAttributes[attr.Key] = anyValueToGo(attr.Value)
				}

				timeStamp := l.TimeUnixNano

				if timeStamp == 0 {
					timeStamp = l.ObservedTimeUnixNano
				}

				log := store.Log{
					SpanID:             hex.EncodeToString(spanID),
					TraceID:            hex.EncodeToString(traceID),
					Service:            service,
					Environment:        environment,
					Attributes:         logAttributes,
					ResourceAttributes: resourceAttributes,
					SeverityText:       l.SeverityText,
					SeverityNumber:     int32(l.SeverityNumber),
					Body:               anyValueToGo(l.Body),
					Timestamp:          time.Unix(0, int64(timeStamp)),
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
