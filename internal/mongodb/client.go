package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoURI = "mongodb://localhost:27017"
	dbName   = "ethereum_monitor"
)

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewClient(ctx context.Context) (*Client, error) {
	credential := options.Credential{
		Username: "root",
		Password: "example",
	}

	mongoOptions := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetAuth(credential).
		SetTimeout(10 * time.Second).
		SetRetryWrites(true)

	client, err := mongo.Connect(ctx, mongoOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}
