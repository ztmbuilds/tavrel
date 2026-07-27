package otlphttp

import (
	"io"
	"net/http"
	"noname/internal/pipeline"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type logExportProcessor interface {
	Process(req *logspb.ExportLogsServiceRequest)
}

func newIngestLogsExportHandler(worker worker, processor logExportProcessor) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		req := &logspb.ExportLogsServiceRequest{}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			writeError(w, r, http.StatusBadRequest, "failed to read request body: "+err.Error(), nil)
			return
		}

		if r.Header.Get("Content-Type") == "application/json" {
			err = protojson.Unmarshal(body, req)
		}
		err = proto.Unmarshal(body, req)

		if err != nil {
			writeError(w, r, http.StatusBadRequest, "failed to unmarshal request body: "+err.Error(), nil)
			return
		}

		if err := worker.Enqueue(pipeline.TaskFunc(func() {
			processor.Process(req)
		})); err != nil {

			writeError(w, r, http.StatusTooManyRequests, err.Error(), map[string]string{
				"Retry-After": "5",
			})
			return
		}

		resp := &logspb.ExportLogsServiceResponse{}
		writeSuccess(w, r, resp)

	}
}
