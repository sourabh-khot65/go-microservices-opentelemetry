package main

import (
	"log"
	"notification-service/internal/api"
	"notification-service/internal/config"
	"notification-service/internal/otel"
)

func main() {
	cfg := config.Load()
	otel.Shutdown = otel.Init(cfg.ServiceName, cfg.OtelEndpoint)
	defer otel.Shutdown()

	handler := api.NewNotificationHandler()
	router := handler.Router()
	log.Printf("Starting notification-service on :%s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}