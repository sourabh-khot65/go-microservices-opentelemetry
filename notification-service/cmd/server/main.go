package main

import (
	"context"
	"log"
	"notification-service/internal/config"
	"notification-service/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create and start server
	srv := server.New(cfg)
	
	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}