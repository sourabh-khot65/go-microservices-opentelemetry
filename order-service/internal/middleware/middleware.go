package middleware

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "The total number of HTTP requests.",
		},
		[]string{"method", "endpoint", "status_code", "service"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "The HTTP request latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code", "service"},
	)

	tracer = otel.Tracer("middleware")
)

// PrometheusMiddleware collects HTTP metrics for Prometheus
func PrometheusMiddleware(serviceName string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
			serviceName,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
			serviceName,
		).Observe(duration)
	})
}

// StructuredLoggingMiddleware provides structured logging for requests
func StructuredLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		slog.Info("HTTP Request",
			"timestamp", param.TimeStamp.Format(time.RFC3339),
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"method", param.Method,
			"path", param.Path,
			"error_message", param.ErrorMessage,
		)
		return ""
	})
}

// TracingMiddleware enhances OpenTelemetry tracing
func TracingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.FullPath())
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.remote_addr", c.ClientIP()),
		)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		if c.Writer.Status() >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	})
}

// ErrorHandlingMiddleware provides consistent error handling
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		slog.Error("Panic recovered",
			"error", recovered,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)

		c.JSON(500, gin.H{
			"error":   "Internal server error",
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		})
	})
}
