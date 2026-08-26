package otlpgrpc

import (
	"context"
	"log/slog"
	"noname/internal/pipeline"

	pb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type worker interface {
	Enqueue(t pipeline.Task) error
}
type traceExportProcessor interface {
	Process(req *pb.ExportTraceServiceRequest)
}
type traceServiceServer struct {
	pb.UnimplementedTraceServiceServer
	processor traceExportProcessor
	worker    worker
}

func newTraceServiceServer(processor traceExportProcessor, worker worker) *traceServiceServer {
	return &traceServiceServer{processor: processor, worker: worker}
}

func (s *traceServiceServer) Export(ctx context.Context, req *pb.ExportTraceServiceRequest) (*pb.ExportTraceServiceResponse, error) {

	if err := s.worker.Enqueue(pipeline.TaskFunc(func() {
		s.processor.Process(req)
	})); err != nil {

		st, err := status.New(codes.Unavailable, "Server is Unavailable").WithDetails(&errdetails.RetryInfo{RetryDelay: &durationpb.Duration{Seconds: 5}})

		if err != nil {
			slog.Error("error returning unavailable response", "error", err)
		}

		return nil, st.Err()
	}

	return nil, nil

}
