package config

import "os"

type Config struct {
	Port         string
	ServiceName  string
	OtelEndpoint string
}

func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8081"),
		ServiceName:  getEnv("SERVICE_NAME", "notification-service"),
		OtelEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
