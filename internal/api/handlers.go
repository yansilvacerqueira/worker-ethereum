// internal/api/handlers.go
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yansilvacerqueira/worker-ethereum/internal/models"
	"github.com/yansilvacerqueira/worker-ethereum/internal/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	mongoClient *mongodb.Client
}

func NewHandler(mongoClient *mongodb.Client) *Handler {
	return &Handler{mongoClient: mongoClient}
}

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := h.mongoClient.Collection("transactions")

	// Parse query parameters
	limit := 100
	skip := 0
	var err error
	if r.URL.Query().Get("limit") != "" {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.URL.Query().Get("skip") != "" {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}

func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := h.mongoClient.Collection("alerts")

	timeRange := time.Hour * 24 // Last 24 hours
	if r.URL.Query().Get("range") != "" {
		hours, _ := strconv.Atoi(r.URL.Query().Get("range"))
		timeRange = time.Hour * time.Duration(hours)
	}

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": time.Now().Add(-timeRange),
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var alerts []models.Alert
	if err := cursor.All(ctx, &alerts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(alerts)
}
