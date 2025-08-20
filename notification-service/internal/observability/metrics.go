package observability

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	requestsTotal        metric.Int64Counter
	requestDuration      metric.Float64Histogram
	requestsInFlight     metric.Int64UpDownCounter
	databaseConnections  metric.Int64UpDownCounter
	databaseOperations   metric.Int64Counter
	businessMetrics      metric.Int64Counter
}

func NewMetrics(serviceName string) (*Metrics, error) {
	meter := otel.Meter(serviceName)

	requestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0, 7.5, 10.0),
	)
	if err != nil {
		return nil, err
	}

	requestsInFlight, err := meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Number of HTTP requests currently being processed"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	databaseConnections, err := meter.Int64UpDownCounter(
		"database_connections_active",
		metric.WithDescription("Number of active database connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return nil, err
	}

	databaseOperations, err := meter.Int64Counter(
		"database_operations_total",
		metric.WithDescription("Total number of database operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, err
	}

	businessMetrics, err := meter.Int64Counter(
		"business_operations_total",
		metric.WithDescription("Total number of business operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		requestsTotal:        requestsTotal,
		requestDuration:      requestDuration,
		requestsInFlight:     requestsInFlight,
		databaseConnections:  databaseConnections,
		databaseOperations:   databaseOperations,
		businessMetrics:      businessMetrics,
	}, nil
}

func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status_code", strconv.Itoa(statusCode)),
		attribute.String("status_class", getStatusClass(statusCode)),
	}

	m.requestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.requestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *Metrics) IncRequestsInFlight(ctx context.Context) {
	m.requestsInFlight.Add(ctx, 1)
}

func (m *Metrics) DecRequestsInFlight(ctx context.Context) {
	m.requestsInFlight.Add(ctx, -1)
}

func (m *Metrics) RecordDatabaseOperation(ctx context.Context, operation, table string, success bool) {
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("table", table),
		attribute.Bool("success", success),
	}
	m.databaseOperations.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *Metrics) SetDatabaseConnections(ctx context.Context, count int64) {
	// Reset and set new value
	m.databaseConnections.Add(ctx, count)
}

func (m *Metrics) RecordBusinessOperation(ctx context.Context, operation string, success bool) {
	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.Bool("success", success),
	}
	m.businessMetrics.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func getStatusClass(statusCode int) string {
	switch {
	case statusCode >= 100 && statusCode < 200:
		return "1xx"
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func IsSuccessStatus(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}