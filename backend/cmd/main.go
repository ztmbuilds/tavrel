package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"noname/internal/config"
	"noname/internal/otlp/otlphttp"
	"noname/internal/pipeline"
	"noname/internal/store"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	srv := otlphttp.NewOTLPHttpServer(sqlLiteStore, worker, cfg)

	go func() {
		err := srv.ListenAndServe()

		//Shutting down the server returns http.ErrServerClosed so we do not want to call log.Fatal on a clean shutdown
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error", "error", err.Error())
			os.Exit(1)
		}
	}()

	logger.Info("OTLP/HTTP server started", "port", cfg.OTLPHTTPPort)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("Forced shutdown:  %w", err)
	}

	worker.Stop()
	logger.Info("Servers stopped cleanly")

	return nil
}
