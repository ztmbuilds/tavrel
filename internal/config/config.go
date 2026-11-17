package config

import (
	"errors"
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
}

func Load() (*Config, error) {

	_ = godotenv.Load(".env")

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
	if retentionDays < 0 {
		return nil, fmt.Errorf("invalid RETENTION_DAYS env: must be >= 0, got %d", retentionDays)
	}
	if retentionDays > maxRetentionDays {
		return nil, fmt.Errorf("invalid RETENTION_DAYS env: must be <= %d, got %d", maxRetentionDays, retentionDays)
	}

	maxRequestBodyBytes, err := strconv.Atoi(getEnv("OTLP_MAX_REQUEST_SIZE", "67108864"))
	if err != nil {
		return nil, fmt.Errorf("failed to load OTLP_MAX_REQUEST_SIZE env: %w", err)
	}

	authToken := getEnv("AUTH_TOKEN", "")
	if authToken != "" {
		if len(authToken) < minAuthTokenLength {
			return nil, fmt.Errorf("failed to load AUTH_TOKEN env: must be at least %d characters", minAuthTokenLength)
		}
		if len(authToken) > maxAuthTokenLength {
			return nil, fmt.Errorf("failed to load AUTH_TOKEN env: must be at most %d characters", maxAuthTokenLength)
		}
		for _, char := range authToken {
			if char < 0x21 || char > 0x7e {
				return nil, errors.New("failed to load AUTH_TOKEN env: must contain only printable ASCII characters without whitespace")
			}
		}
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
	}

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
