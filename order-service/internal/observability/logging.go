package observability

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// SetupStructuredLogging configures structured logging with OpenTelemetry correlation
func SetupStructuredLogging(serviceName, environment string) {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// Use JSON formatting in production, text in development
	var handler slog.Handler
	if environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Wrap with correlation handler
	correlationHandler := &CorrelationHandler{
		Handler:     handler,
		serviceName: serviceName,
	}

	slog.SetDefault(slog.New(correlationHandler))
}

// CorrelationHandler adds trace correlation to log entries
type CorrelationHandler struct {
	slog.Handler
	serviceName string
}

func (h *CorrelationHandler) Handle(ctx context.Context, record slog.Record) error {
	// Add service name
	record.AddAttrs(slog.String("service", h.serviceName))

	// Add trace correlation if available
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		record.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		record.AddAttrs(slog.String("request_id", requestID))
	}

	return h.Handler.Handle(ctx, record)
}

// LogWithContext is a helper function for contextual logging
func LogWithContext(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	logger := slog.Default()
	logger.Log(ctx, level, msg, args...)
}

// InfoWithContext logs at info level with context
func InfoWithContext(ctx context.Context, msg string, args ...interface{}) {
	LogWithContext(ctx, slog.LevelInfo, msg, args...)
}

// ErrorWithContext logs at error level with context
func ErrorWithContext(ctx context.Context, msg string, args ...interface{}) {
	LogWithContext(ctx, slog.LevelError, msg, args...)
}

// WarnWithContext logs at warn level with context
func WarnWithContext(ctx context.Context, msg string, args ...interface{}) {
	LogWithContext(ctx, slog.LevelWarn, msg, args...)
}

// DebugWithContext logs at debug level with context
func DebugWithContext(ctx context.Context, msg string, args ...interface{}) {
	LogWithContext(ctx, slog.LevelDebug, msg, args...)
}