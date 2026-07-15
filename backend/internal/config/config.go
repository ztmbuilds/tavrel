package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath string
}

func Load() *Config {

	_ = godotenv.Load()
	cfg := &Config{
		DBPath: getEnv("DB_PATH", "store.db"),
	}

	return cfg

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
