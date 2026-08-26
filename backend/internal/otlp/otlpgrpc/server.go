package otlpgrpc

import (
	"context"
	"log/slog"
	"time"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	tracespb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func loggingUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	slog.Info("OTLP/gRPC request",
		"method", info.FullMethod,
		"status", status.Code(err).String(),
		"duration", time.Since(start).String(),
	)

	return resp, err
}

func NewOTLPGRPCServer(worker worker, traceProcessor traceExportProcessor, logsProcessor logExportProcessor) *grpc.Server {

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MiB
		grpc.MaxSendMsgSize(4*1024*1024),
		grpc.UnaryInterceptor(loggingUnaryInterceptor))

	tracespb.RegisterTraceServiceServer(grpcServer, newTraceServiceServer(traceProcessor, worker))
	logspb.RegisterLogsServiceServer(grpcServer, newLogsServiceServer(logsProcessor, worker))

	reflection.Register(grpcServer)

	return grpcServer
}
