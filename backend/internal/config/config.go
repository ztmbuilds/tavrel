package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath              string
	OTLPHTTPPort        string
	OTLPGRPCPort        string
	WorkerCount         int
	TaskQueueSize       int
	AppPort             string
	RetentionDays       int
	MaxRequestBodyBytes int
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

	retentionDays, err := strconv.Atoi(getEnv("RETENTION_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("failed to load RETENTION_DAYS env: %w", err)
	}

	maxRequestBodyBytes, err := strconv.Atoi(getEnv("OTLP_MAX_REQUEST_SIZE", "67108864"))
	if err != nil {
		return nil, fmt.Errorf("failed to load OTLP_MAX_REQUEST_SIZE env: %w", err)
	}

	cfg := &Config{
		DBPath:              getEnv("DB_PATH", "store.db"),
		OTLPHTTPPort:        getEnv("OTLP_HTTP_PORT", "4318"),
		OTLPGRPCPort:        getEnv("OTLP_GRPC_PORT", "4317"),
		WorkerCount:         workerCount,
		TaskQueueSize:       taskQueueSize,
		AppPort:             getEnv("APP_PORT", "5678"),
		RetentionDays:       retentionDays,
		MaxRequestBodyBytes: maxRequestBodyBytes,
	}

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
