package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	// Server configuration
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Database configuration
	DBDsn          string
	DBMaxOpenConns int
	DBMaxIdleConns int
	DBConnMaxLife  time.Duration
	DBConnMaxIdle  time.Duration

	// OpenTelemetry configuration
	OtelEndpoint   string
	ServiceName    string
	ServiceVersion string
	Environment    string
	SampleRate     float64

	// Logging configuration
	LogLevel  string
	LogFormat string // json or text

	// Health check configuration
	HealthCheckPath string
	ReadinessPath   string
	MetricsPath     string

	// Feature flags
	EnableMetrics bool
	EnableTracing bool
	EnableLogging bool
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		// Server
		Port:         getEnv("PORT", "8080"),
		Host:         getEnv("HOST", "0.0.0.0"),
		ReadTimeout:  getDuration("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getDuration("WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getDuration("IDLE_TIMEOUT", 120*time.Second),

		// Database
		DBDsn:          getEnv("DB_DSN", "postgres://demo:demo@localhost:5432/demo?sslmode=disable"),
		DBMaxOpenConns: getInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns: getInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLife:  getDuration("DB_CONN_MAX_LIFE", 5*time.Minute),
		DBConnMaxIdle:  getDuration("DB_CONN_MAX_IDLE", 5*time.Minute),

		// OpenTelemetry
		OtelEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		ServiceName:    getEnv("SERVICE_NAME", "order-service"),
		ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		SampleRate:     getFloat64("OTEL_SAMPLE_RATE", 1.0),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		// Health checks
		HealthCheckPath: getEnv("HEALTH_CHECK_PATH", "/health"),
		ReadinessPath:   getEnv("READINESS_PATH", "/ready"),

		// Feature flags
		EnableMetrics: getBool("ENABLE_METRICS", true),
		EnableTracing: getBool("ENABLE_TRACING", true),
		EnableLogging: getBool("ENABLE_LOGGING", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
