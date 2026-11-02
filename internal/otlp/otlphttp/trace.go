package otlphttp

import (
	"io"
	"net/http"
	"noname/internal/pipeline"

	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type worker interface {
	Enqueue(t pipeline.Task) error
}

type traceExportProcessor interface {
	Process(req *tracepb.ExportTraceServiceRequest)
}

func newIngestTraceExportHandler(worker worker, processor traceExportProcessor) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		req := &tracepb.ExportTraceServiceRequest{}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			writeError(w, r, http.StatusBadRequest, "failed to read request body: "+err.Error(), nil)
			return
		}

		if r.Header.Get("Content-Type") == "application/json" {
			err = protojson.Unmarshal(body, req)

		} else {
			err = proto.Unmarshal(body, req)

		}

		if err != nil {
			writeError(w, r, http.StatusBadRequest, "failed to unmarshal request body: "+err.Error(), nil)
			return
		}

		if err := worker.Enqueue(pipeline.TaskFunc(func() {
			processor.Process(req)
		})); err != nil {

			//buffer is full
			writeError(w, r,
				http.StatusTooManyRequests,
				err.Error(),
				map[string]string{
					"Retry-After": "5",
				})
			return
		}

		resp := &tracepb.ExportTraceServiceResponse{}

		writeSuccess(w, r, resp)

	}

}
