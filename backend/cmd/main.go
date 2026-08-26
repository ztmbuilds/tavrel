package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"noname/internal/api"
	"noname/internal/config"
	"noname/internal/cron"
	"noname/internal/otlp/otlpgrpc"
	"noname/internal/otlp/otlphttp"
	"noname/internal/pipeline"
	"noname/internal/store"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()

	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	sqlLiteStore, err := store.NewSQLiteStore(cfg.DBPath)

	if err != nil {

		return err
	}

	defer sqlLiteStore.Close()

	worker := pipeline.NewWorker(cfg.TaskQueueSize, cfg.WorkerCount)
	worker.Start()

	retentioncron := cron.New(cfg.RetentionDays, sqlLiteStore.DeleteOldData)
	retentioncron.Start(context.Background())
	defer retentioncron.Stop()

	traceProcessor := pipeline.NewTraceExportProcessor(sqlLiteStore)
	logsProcessor := pipeline.NewLogsExportProcessor(sqlLiteStore)

	otlphttpSrv := otlphttp.NewOTLPHttpServer(worker, traceProcessor, logsProcessor, cfg)
	otlpgrpcSrv := otlpgrpc.NewOTLPGRPCServer(worker, traceProcessor, logsProcessor)

	appSrv := api.NewAPPHttpServer(sqlLiteStore, cfg)

	go startHttpServer(otlphttpSrv, "OTLP/HTTP", logger)
	go startHttpServer(appSrv, "APP", logger)

	go startgRPCServer(otlpgrpcSrv, cfg.OTLPGRPCPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var shutdownErr error
	if err := otlphttpSrv.Shutdown(ctx); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("otlp server forced shutdown: %w", err))
	}
	if err := appSrv.Shutdown(ctx); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("%w; api server forced shutdown: %w", shutdownErr, err))

	}

	otlpgrpcSrv.GracefulStop()

	worker.Stop()
	logger.Info("Servers stopped cleanly")

	return shutdownErr
}

func startHttpServer(srv *http.Server, srvName string, logger *slog.Logger) {
	logger.Info(fmt.Sprintf("%s server starting", srvName), "port", srv.Addr)

	err := srv.ListenAndServe()

	// Shutting down the server returns http.ErrServerClosed, so we don't
	// want to treat a clean shutdown as a fatal error.
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "server", srvName, "error", err)
		os.Exit(1)
	}
}

func startgRPCServer(srv *grpc.Server, port string) {

	addr := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		slog.Error("failed to listen", "err", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("gRPC server listening on %s", addr))

	if err := srv.Serve(lis); err != nil {
		slog.Error("failed to serve", "err", err)
		os.Exit(1)
	}

}
