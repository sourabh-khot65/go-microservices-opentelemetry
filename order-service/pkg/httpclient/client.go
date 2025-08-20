package httpclient

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// New creates a new HTTP client with OpenTelemetry instrumentation
func New() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

// NewWithTimeout creates a new HTTP client with custom timeout and OpenTelemetry instrumentation
func NewWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
