package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Hash         string             `bson:"hash"`
	From         string             `bson:"from"`
	To           string             `bson:"to"`
	Value        float64            `bson:"value"`
	TokenName    string             `bson:"token_name,omitempty"`
	TokenValue   float64            `bson:"token_value,omitempty"`
	IsHighValue  bool               `bson:"is_high_value"`
	IsSuspicious bool               `bson:"is_suspicious"`
	AlertType    string             `bson:"alert_type,omitempty"`
	Timestamp    time.Time          `bson:"timestamp"`
}

type Alert struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	TxHash      string             `bson:"tx_hash"`
	AlertType   string             `bson:"alert_type"`
	Description string             `bson:"description"`
	Timestamp   time.Time          `bson:"timestamp"`
}
