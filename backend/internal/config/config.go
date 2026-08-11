package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath        string
	OTLPHTTPPort  string
	WorkerCount   int
	TaskQueueSize int
	AppPort       string
}

func Load() (*Config, error) {

	_ = godotenv.Load()

	workerCount, err := strconv.Atoi(getEnv("WORKER_COUNT", "4"))

	if err != nil {
		return nil, fmt.Errorf("failed to load WORKER_COUNT env: %w", err)
	}

	taskQueueSize, err := strconv.Atoi(getEnv("TASK_QUEUE_SIZE", "5000"))

	if err != nil {
		return nil, fmt.Errorf("failed to load TASK_QUEUE_SIZE env: %w", err)
	}

	cfg := &Config{
		DBPath:        getEnv("DB_PATH", "store.db"),
		OTLPHTTPPort:  getEnv("OTLP_HTTP_PORT", "4318"),
		WorkerCount:   workerCount,
		TaskQueueSize: taskQueueSize,
		AppPort:       getEnv("APP_PORT", "5678"),
	}

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
