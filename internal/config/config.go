package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerPort      string
	DatabaseURL     string
	Environment     string
	LogLevel        string
	OTELEndpoint    string
	OTELServiceName string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		Environment:     getEnv("ENVIRONMENT", "dev"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		OTELEndpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OTELServiceName: getEnv("OTEL_SERVICE_NAME", "go-crud-service"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	if _, err := strconv.Atoi(cfg.ServerPort); err != nil {
		return nil, fmt.Errorf("SERVER_PORT must be a valid integer, got: %s", cfg.ServerPort)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
