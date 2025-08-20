package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/api"
	"notification-service/internal/config"
	"notification-service/internal/observability"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config     *config.Config
	httpServer *http.Server
	telemetry  *observability.SimpleTelemetry
	metrics    *observability.Metrics
}

func New(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Setup structured logging
	observability.SetupStructuredLogging(s.config.ServiceName, s.config.Environment)

	// Initialize telemetry
	telConfig := observability.TelemetryConfig{
		ServiceName:    s.config.ServiceName,
		ServiceVersion: s.config.ServiceVersion,
		OtelEndpoint:   s.config.OtelEndpoint,
		Environment:    s.config.Environment,
	}

	var err error
	s.telemetry, err = observability.NewSimpleTelemetry(telConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Initialize metrics
	s.metrics, err = observability.NewMetrics(s.config.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize HTTP server
	s.initHTTPServer()

	// Start server
	go func() {
		slog.InfoContext(ctx, "Starting HTTP server",
			"address", fmt.Sprintf("%s:%s", s.config.Host, s.config.Port),
			"service", s.config.ServiceName,
		)

		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "HTTP server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	return s.waitForShutdown(ctx)
}

func (s *Server) initHTTPServer() {
	// Set Gin mode based on environment
	if s.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Initialize middleware
	middleware := observability.NewMiddleware(s.config.ServiceName, s.metrics)

	// Add core middleware
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	if s.config.EnableTracing {
		router.Use(middleware.TracingMiddleware())
	}

	if s.config.EnableMetrics {
		router.Use(middleware.MetricsMiddleware())
	}

	router.Use(middleware.StructuredLoggingMiddleware())

	// Health check endpoints
	router.GET(s.config.HealthCheckPath, middleware.HealthCheckHandler())
	router.GET(s.config.ReadinessPath, middleware.ReadinessHandler())

	// Metrics endpoint
	if s.config.EnableMetrics {
		router.GET(s.config.MetricsPath, middleware.MetricsHandler())
	}

	// Business logic setup
	handler := api.NewNotificationHandler()
	handler.RegisterRoutes(router)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%s", s.config.Host, s.config.Port),
		Handler:        router,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}

func (s *Server) waitForShutdown(ctx context.Context) error {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-quit
	slog.InfoContext(ctx, "Received shutdown signal", "signal", sig.String())

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.shutdown(shutdownCtx)
}

func (s *Server) shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting graceful shutdown")

	var shutdownErrors []error

	// Shutdown HTTP server
	if s.httpServer != nil {
		slog.InfoContext(ctx, "Shutting down HTTP server")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("HTTP server shutdown error: %w", err))
		}
	}

	// Shutdown telemetry
	if s.telemetry != nil {
		slog.InfoContext(ctx, "Shutting down telemetry")
		if err := s.telemetry.Shutdown(ctx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("telemetry shutdown error: %w", err))
		}
	}

	if len(shutdownErrors) > 0 {
		slog.ErrorContext(ctx, "Shutdown completed with errors", "errors", shutdownErrors)
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}

	slog.InfoContext(ctx, "Graceful shutdown completed successfully")
	return nil
}