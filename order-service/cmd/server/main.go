package main

import (
	"context"
	"database/sql"
	"log"
	"order-service/internal/api"
	"order-service/internal/config"
	"order-service/internal/otel"
	"order-service/internal/repository"
	"order-service/internal/service"

	"github.com/XSAM/otelsql"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	shutdown, err := otel.Init(ctx, cfg.ServiceName, cfg.OtelEndpoint)
	if err != nil {
		log.Fatalf("failed to init otel: %v", err)
	}
	defer shutdown()

	driverName, err := otelsql.Register("pgx", otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		log.Fatalf("failed to register otel pgx: %v", err)
	}
	db, err := sql.Open(driverName, cfg.DBDsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo)
	handler := api.NewOrderHandler(svc)

	r := gin.Default()
	handler.RegisterRoutes(r)

	log.Printf("order-service running on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
