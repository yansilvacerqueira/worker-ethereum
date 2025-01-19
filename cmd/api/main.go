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

var (
	MongoURI      = "mongodb://localhost:27017"
	MongoUser     = "root"
	MongoPassword = "example"
	DatabaseName  = "ethereum_monitor"
)

func main() {
	ctx := context.Background()

	mongoClient, err := mongodb.NewClient(ctx, DatabaseName, MongoUser, MongoPassword, DatabaseName)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Close(ctx)

	handler := api.NewHandler(mongoClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/transactions", handler.GetTransactions)
	mux.HandleFunc("/api/alerts", handler.GetAlerts)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

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
