package otlpgrpc

import (
	"context"
	"log/slog"
	"noname/internal/config"
	"time"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	tracespb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"

	_ "google.golang.org/grpc/encoding/gzip"
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

func NewOTLPGRPCServer(worker worker, traceProcessor traceExportProcessor, logsProcessor logExportProcessor, cfg *config.Config) *grpc.Server {

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.MaxRequestBodyBytes),
		grpc.MaxSendMsgSize(cfg.MaxRequestBodyBytes),
		grpc.UnaryInterceptor(loggingUnaryInterceptor))

	tracespb.RegisterTraceServiceServer(grpcServer, newTraceServiceServer(traceProcessor, worker))
	logspb.RegisterLogsServiceServer(grpcServer, newLogsServiceServer(logsProcessor, worker))

	reflection.Register(grpcServer)

	return grpcServer
}
