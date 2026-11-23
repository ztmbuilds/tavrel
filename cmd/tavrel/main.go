package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ztmbuilds/tavrel/internal/api"
	"github.com/ztmbuilds/tavrel/internal/config"
	"github.com/ztmbuilds/tavrel/internal/cron"
	"github.com/ztmbuilds/tavrel/internal/otlp/otlpgrpc"
	"github.com/ztmbuilds/tavrel/internal/otlp/otlphttp"
	"github.com/ztmbuilds/tavrel/internal/pipeline"
	"github.com/ztmbuilds/tavrel/internal/store"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := healthcheck("127.0.0.1:" + cfg.AppPort); err != nil {
			slog.Error("Healthcheck failed", "error", err)
			os.Exit(1)
		}

		return
	}

	if err := run(cfg); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if cfg.AuthToken == "" {
		logger.Warn("Authentication disabled")
	} else {
		logger.Info("Authentication enabled")
	}

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
	otlpgrpcSrv := otlpgrpc.NewOTLPGRPCServer(worker, traceProcessor, logsProcessor, cfg)
	appSrv := api.NewAPPHttpServer(sqlLiteStore, cfg)

	// Bind all listeners before starting servers.
	appListener, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		return fmt.Errorf("listening on App port: %w", err)
	}
	defer appListener.Close()
	logger.Info("APP server listening", "addr", appListener.Addr())

	otlpHTTPListener, err := net.Listen("tcp", ":"+cfg.OTLPHTTPPort)
	if err != nil {
		return fmt.Errorf("listening on OTLP/HTTP port: %w", err)
	}
	defer otlpHTTPListener.Close()
	logger.Info("OTLP/HTTP server listening", "addr", otlpHTTPListener.Addr())

	otlpGRPCListener, err := net.Listen("tcp", ":"+cfg.OTLPGRPCPort)
	if err != nil {
		return fmt.Errorf("listening on OTLP/gRPC port: %w", err)
	}
	defer otlpGRPCListener.Close()
	logger.Info("OTLP/GRPC server listening", "addr", otlpGRPCListener.Addr())

	errCh := make(chan error, 3)
	go func() {
		errCh <- startHttpServer(otlphttpSrv, otlpHTTPListener)
	}()

	go func() {
		errCh <- startHttpServer(appSrv, appListener)
	}()

	go func() {
		errCh <- startgRPCServer(otlpgrpcSrv, otlpGRPCListener)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	var runErr error

	select {
	case err := <-errCh:
		if err != nil {
			runErr = fmt.Errorf("server failed: %w", err)
		} else {
			runErr = errors.New("server stopped unexpectedly")
		}

	case <-quit:
		logger.Info("Shutdown signal received")
	}

	logger.Info("Shutting down servers")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var shutdownErr error
	if err := otlphttpSrv.Shutdown(ctx); err != nil {
		shutdownErr = errors.Join(
			shutdownErr,
			fmt.Errorf("OTLP/HTTP server shutdown: %w", err),
		)
	}

	if err := appSrv.Shutdown(ctx); err != nil {
		shutdownErr = errors.Join(
			shutdownErr,
			fmt.Errorf("APP server shutdown: %w", err),
		)
	}

	grpcStopped := make(chan struct{})

	go func() {
		otlpgrpcSrv.GracefulStop()
		close(grpcStopped)
	}()

	select {
	case <-grpcStopped:
		logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		logger.Warn("gRPC graceful shutdown timed out")
		otlpgrpcSrv.Stop()
	}

	worker.Stop()
	logger.Info("Servers stopped cleanly")

	return errors.Join(runErr, shutdownErr)
}

func startHttpServer(srv *http.Server, lis net.Listener) error {
	err := srv.Serve(lis)

	// Shutting down the server returns http.ErrServerClosed, so we don't
	// want to treat a clean shutdown as a fatal error.
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func startgRPCServer(srv *grpc.Server, lis net.Listener) error {

	return srv.Serve(lis)
}
