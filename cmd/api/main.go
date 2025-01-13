package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yansilvacerqueira/worker-ethereum/internal/api"
	"github.com/yansilvacerqueira/worker-ethereum/internal/mongodb"
)

func main() {
	ctx := context.Background()

	// Initialize MongoDB client
	mongoClient, err := mongodb.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Close(ctx)

	// Initialize handler
	handler := api.NewHandler(mongoClient)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions", handler.GetTransactions)
	mux.HandleFunc("/api/alerts", handler.GetAlerts)

	// Configure server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
