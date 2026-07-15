package main

import (
	"log/slog"
	"noname/internal/config"
	"noname/internal/store"
	"os"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application startup failed", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	sqlLiteStore, err := store.NewSQLiteStore(cfg.DBPath)

	if err != nil {

		return err
	}

	defer sqlLiteStore.Close()

	return nil
}
