package observability

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Middleware struct {
	serviceName string
	metrics     *Metrics
	tracer      trace.Tracer
}

func NewMiddleware(serviceName string, metrics *Metrics) *Middleware {
	return &Middleware{
		serviceName: serviceName,
		metrics:     metrics,
		tracer:      otel.Tracer(serviceName),
	}
}

// Note: Metrics are exported via OTLP to OpenTelemetry Collector
// No Prometheus scraping endpoint needed

// HealthCheckHandler returns a health check endpoint
func (m *Middleware) HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   m.serviceName,
			"timestamp": time.Now().UTC(),
		})
	}
}

// ReadinessHandler returns a readiness probe endpoint
func (m *Middleware) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add actual readiness checks here (database connectivity, etc.)
		c.JSON(200, gin.H{
			"status":    "ready",
			"service":   m.serviceName,
			"timestamp": time.Now().UTC(),
		})
	}
}

// TracingMiddleware adds OpenTelemetry tracing with additional context
func (m *Middleware) TracingMiddleware() gin.HandlerFunc {
	return otelgin.Middleware(m.serviceName)
}

// MetricsMiddleware collects HTTP metrics
func (m *Middleware) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Increment in-flight requests
		m.metrics.IncRequestsInFlight(c.Request.Context())
		defer m.metrics.DecRequestsInFlight(c.Request.Context())

		c.Next()

		duration := time.Since(start)
		m.metrics.RecordHTTPRequest(
			c.Request.Context(),
			c.Request.Method,
			path,
			c.Writer.Status(),
			duration,
		)
	}
}

// StructuredLoggingMiddleware provides structured request logging with correlation
func (m *Middleware) StructuredLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Trace correlation is automatically handled by CorrelationHandler in logging.go

		slog.InfoContext(param.Request.Context(), "HTTP Request",
			"timestamp", param.TimeStamp.Format(time.RFC3339),
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency.String(),
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"bytes_in", param.Request.ContentLength,
			"bytes_out", param.BodySize,
			"service", m.serviceName,
		)
		return ""
	})
}

// ErrorHandlingMiddleware provides comprehensive error handling and recovery
func (m *Middleware) ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		span := trace.SpanFromContext(c.Request.Context())
		span.SetAttributes(
			attribute.Bool("error", true),
			attribute.String("error.type", "panic"),
		)

		slog.ErrorContext(c.Request.Context(), "Panic recovered",
			"error", recovered,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"service", m.serviceName,
		)

		c.JSON(500, gin.H{
			"error":      "Internal server error",
			"code":       "INTERNAL_ERROR",
			"message":    "An unexpected error occurred",
			"request_id": GetRequestID(c.Request.Context()),
		})
	})
}

// RequestIDMiddleware adds request ID for tracing
func (m *Middleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
				requestID = span.SpanContext().TraceID().String()
			}
		}

		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}
