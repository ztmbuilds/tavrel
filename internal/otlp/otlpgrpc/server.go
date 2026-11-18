package otlpgrpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/ztmbuilds/tavrel/internal/config"
	"github.com/ztmbuilds/tavrel/internal/otlp"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	tracespb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func failureLoggingUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	if err != nil {
		slog.Warn("OTLP/gRPC request failed",
			"method", info.FullMethod,
			"status", status.Code(err).String(),
			"duration", time.Since(start).String(),
		)
	}

	return resp, err
}

func authUnaryInterceptor(token string) grpc.UnaryServerInterceptor {
	if token == "" {
		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}

		for _, value := range md.Get("x-api-key") {
			if otlp.TokensEqual(token, value) {
				return handler(ctx, req)
			}
		}

		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
}

func NewOTLPGRPCServer(worker worker, traceProcessor traceExportProcessor, logsProcessor logExportProcessor, cfg *config.Config) *grpc.Server {

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.MaxRequestBodyBytes),
		grpc.MaxSendMsgSize(cfg.MaxRequestBodyBytes),
		grpc.ChainUnaryInterceptor(
			failureLoggingUnaryInterceptor,
			authUnaryInterceptor(cfg.IngestToken),
		))

	tracespb.RegisterTraceServiceServer(grpcServer, newTraceServiceServer(traceProcessor, worker))
	logspb.RegisterLogsServiceServer(grpcServer, newLogsServiceServer(logsProcessor, worker))

	reflection.Register(grpcServer)

	return grpcServer
}
