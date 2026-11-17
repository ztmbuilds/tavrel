package otlpgrpc

import (
	"context"
	"log/slog"
	"github.com/ztmbuilds/crimp/internal/pipeline"

	pb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type logExportProcessor interface {
	Process(req *pb.ExportLogsServiceRequest)
}

type logsServiceServer struct {
	pb.UnimplementedLogsServiceServer
	processor logExportProcessor
	worker    worker
}

func newLogsServiceServer(processor logExportProcessor, worker worker) *logsServiceServer {
	return &logsServiceServer{processor: processor, worker: worker}
}

func (s *logsServiceServer) Export(ctx context.Context, req *pb.ExportLogsServiceRequest) (*pb.ExportLogsServiceResponse, error) {

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
