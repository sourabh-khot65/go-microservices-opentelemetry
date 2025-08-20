package config

import "os"

type Config struct {
	Port         string
	ServiceName  string
	OtelEndpoint string
	DBDsn        string
}

func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8080"),
		ServiceName:  getEnv("SERVICE_NAME", "order-service"),
		OtelEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		DBDsn:        getEnv("DB_DSN", "postgres://demo:demo@localhost:5432/demo?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
