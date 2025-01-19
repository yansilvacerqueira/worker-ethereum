package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/yansilvacerqueira/worker-ethereum/internal/models"
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

func NewClient(ctx context.Context, uri, username, password, dbName string) (*Client, error) {
	credential := options.Credential{
		Username: username,
		Password: password,
	}

	clientOpts := options.Client().
		ApplyURI(uri).
		SetAuth(credential).
		SetTimeout(10 * time.Second).
		SetRetryWrites(true)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

func (c *Client) SaveTransaction(ctx context.Context, tx models.Transaction) error {
	_, err := c.db.Collection("transactions").InsertOne(ctx, tx)
	return err
}

func (c *Client) SaveAlert(ctx context.Context, alert models.Alert) error {
	_, err := c.db.Collection("alerts").InsertOne(ctx, alert)
	return err
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}
