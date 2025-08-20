package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/internal/api"
	"order-service/internal/config"
	"order-service/internal/observability"
	"order-service/internal/repository"
	"order-service/internal/services"

	"github.com/XSAM/otelsql"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Server struct {
	config     *config.Config
	httpServer *http.Server
	db         *sql.DB
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

	// Initialize database
	if err := s.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
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

func (s *Server) initDatabase() error {
	// Register instrumented driver
	driverName, err := otelsql.Register("pgx", otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return fmt.Errorf("failed to register otel pgx: %w", err)
	}

	// Open database connection
	s.db, err = sql.Open(driverName, s.config.DBDsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	s.db.SetMaxOpenConns(s.config.DBMaxOpenConns)
	s.db.SetMaxIdleConns(s.config.DBMaxIdleConns)
	s.db.SetConnMaxLifetime(s.config.DBConnMaxLife)
	s.db.SetConnMaxIdleTime(s.config.DBConnMaxIdle)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
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

	// Business logic setup
	orderRepo := repository.NewOrderRepository(s.db)
	productRepo := repository.NewProductRepository(s.db)
	notificationService := services.NewNotificationService(s.config.NotificationServiceURL)
	
	// Order handler
	orderHandler := api.NewOrderHandler(orderRepo, productRepo, notificationService)
	// Product handler
	productHandler := api.NewProductHandler(productRepo)

	// Register business routes
	router.POST("/orders", orderHandler.CreateOrder)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.GET("/orders", orderHandler.GetAllOrders)
	
	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products/:id", productHandler.GetProduct)
	router.GET("/products", productHandler.GetAllProducts)

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

	// Close database connections
	if s.db != nil {
		slog.InfoContext(ctx, "Closing database connections")
		if err := s.db.Close(); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("database close error: %w", err))
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
