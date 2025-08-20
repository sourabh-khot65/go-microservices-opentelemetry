package main

import (
	"context"
	"log"
	"order-service/internal/config"
	"order-service/internal/server"
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
