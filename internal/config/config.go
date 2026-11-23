package config

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// maxRetentionDays is the largest number of days that fits in a time.Duration
// without overflowing when multiplied by 24h.
const maxRetentionDays = int(math.MaxInt64 / int64(24*time.Hour))

const (
	minAuthTokenLength = 16
	maxAuthTokenLength = 512
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
	AuthToken           string
	IngestToken         string
}

func Load() (*Config, error) {

	_ = godotenv.Load(".env")

	workerCount, err := strconv.Atoi(getEnv("WORKER_COUNT", "4"))

	if err != nil {
		return nil, fmt.Errorf("failed to load WORKER_COUNT env: %w", err)
	}

	taskQueueSize, err := strconv.Atoi(getEnv("TASK_QUEUE_SIZE", "256"))

	if err != nil {
		return nil, fmt.Errorf("failed to load TASK_QUEUE_SIZE env: %w", err)
	}

	retentionDays, err := strconv.Atoi(getEnv("RETENTION_DAYS", "30"))
	if err != nil {
		return nil, fmt.Errorf("failed to load RETENTION_DAYS env: %w", err)
	}
	if retentionDays < 0 {
		return nil, fmt.Errorf("invalid RETENTION_DAYS env: must be >= 0, got %d", retentionDays)
	}
	if retentionDays > maxRetentionDays {
		return nil, fmt.Errorf("invalid RETENTION_DAYS env: must be <= %d, got %d", maxRetentionDays, retentionDays)
	}

	maxRequestBodyBytes, err := strconv.Atoi(getEnv("OTLP_MAX_REQUEST_SIZE", "8388608"))
	if err != nil {
		return nil, fmt.Errorf("failed to load OTLP_MAX_REQUEST_SIZE env: %w", err)
	}

	authToken := getEnv("AUTH_TOKEN", "")
	if err := validateTokenEnv("AUTH_TOKEN", authToken); err != nil {
		return nil, err
	}

	ingestToken := getEnv("INGEST_TOKEN", "")
	if err := validateTokenEnv("INGEST_TOKEN", ingestToken); err != nil {
		return nil, err
	}

	cfg := &Config{
		DBPath:              getEnv("DB_PATH", "data/store.db"),
		OTLPHTTPPort:        getEnv("OTLP_HTTP_PORT", "4318"),
		OTLPGRPCPort:        getEnv("OTLP_GRPC_PORT", "4317"),
		WorkerCount:         workerCount,
		TaskQueueSize:       taskQueueSize,
		AppPort:             getEnv("APP_PORT", "5678"),
		RetentionDays:       retentionDays,
		MaxRequestBodyBytes: maxRequestBodyBytes,
		AuthToken:           authToken,
		IngestToken:         ingestToken,
	}

	return cfg, nil

}

func validateTokenEnv(name, token string) error {
	if token == "" {
		return nil
	}

	if len(token) < minAuthTokenLength {
		return fmt.Errorf("failed to load %s env: must be at least %d characters", name, minAuthTokenLength)
	}
	if len(token) > maxAuthTokenLength {
		return fmt.Errorf("failed to load %s env: must be at most %d characters", name, maxAuthTokenLength)
	}

	for _, char := range token {
		if char < 0x21 || char > 0x7e {
			return fmt.Errorf("failed to load %s env: must contain only printable ASCII characters without whitespace", name)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
