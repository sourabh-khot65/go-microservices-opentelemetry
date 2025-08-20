package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	OtelEndpoint   string
	Environment    string
}

type Telemetry struct {
	traceProvider  *trace.TracerProvider
	metricProvider *metric.MeterProvider
	logProvider    *log.LoggerProvider
	resource       *resource.Resource
}

func NewTelemetry(config TelemetryConfig) (*Telemetry, error) {
	// Create resource with comprehensive service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
			semconv.ServiceInstanceID(fmt.Sprintf("%s-%d", config.ServiceName, os.Getpid())),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tel := &Telemetry{
		resource: res,
	}

	// Initialize tracing
	if err := tel.initTracing(config.OtelEndpoint); err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize metrics
	if err := tel.initMetrics(config.OtelEndpoint); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize logging
	if err := tel.initLogging(config.OtelEndpoint); err != nil {
		return nil, fmt.Errorf("failed to initialize logging: %w", err)
	}

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("Telemetry initialized successfully", 
		"service", config.ServiceName,
		"endpoint", config.OtelEndpoint,
		"environment", config.Environment,
	)

	return tel, nil
}

func (t *Telemetry) initTracing(endpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	t.traceProvider = trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(t.resource),
		trace.WithSampler(trace.AlwaysSample()),
	)

	otel.SetTracerProvider(t.traceProvider)
	return nil
}

func (t *Telemetry) initMetrics(endpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection for metrics: %w", err)
	}

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Also create Prometheus exporter for direct scraping
	prometheusExporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	t.metricProvider = metric.NewMeterProvider(
		metric.WithResource(t.resource),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(15*time.Second),
		)),
		metric.WithReader(prometheusExporter),
	)

	otel.SetMeterProvider(t.metricProvider)
	return nil
}

func (t *Telemetry) initLogging(endpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection for logs: %w", err)
	}

	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	t.logProvider = log.NewLoggerProvider(
		log.WithResource(t.resource),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	global.SetLoggerProvider(t.logProvider)
	return nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error

	if t.traceProvider != nil {
		if err := t.traceProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("trace provider shutdown: %w", err))
		}
	}

	if t.metricProvider != nil {
		if err := t.metricProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("metric provider shutdown: %w", err))
		}
	}

	if t.logProvider != nil {
		if err := t.logProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("log provider shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}