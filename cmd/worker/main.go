package main

import (
	"context"
	"log"

	"github.com/yansilvacerqueira/worker-ethereum/internal/config"
	"github.com/yansilvacerqueira/worker-ethereum/internal/monitor"
)

func main() {
	ctx := context.Background()

	cfg := config.NewDefaultConfig()

	mon, err := monitor.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create monitor: %v", err)
	}

	log.Println("Starting Ethereum transaction monitor...")
	if err := mon.Start(ctx); err != nil {
		log.Fatalf("Monitor failed: %v", err)
	}
}
