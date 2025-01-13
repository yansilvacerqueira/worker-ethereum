package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/time/rate"
)

type Config struct {
	EthereumNode       string
	HighValueThreshold float64
	WatchedTokens      []common.Address
	WatchedContracts   []common.Address
	MongoURI           string
	MongoUser          string
	MongoPassword      string
	DatabaseName       string
}

type Transaction struct {
	ID           primitive.ObjectID `bson:"_id"`
	Hash         string             `bson:"hash"`
	From         string             `bson:"from"`
	To           string             `bson:"to"`
	Value        float64            `bson:"value"`
	IsHighValue  bool               `bson:"is_high_value"`
	IsSuspicious bool               `bson:"is_suspicious"`
	AlertType    string             `bson:"alert_type,omitempty"`
	Timestamp    time.Time          `bson:"timestamp"`
}

type Alert struct {
	ID          primitive.ObjectID `bson:"_id"`
	Type        string             `bson:"type"`
	Description string             `bson:"description"`
	TxHash      string             `bson:"tx_hash"`
	Timestamp   time.Time          `bson:"timestamp"`
}

type TransactionMonitor struct {
	config    Config
	alertChan chan Alert
	client    *ethclient.Client
	db        *mongo.Database
	txCache   *sync.Map
	limiter   *rate.Limiter
}

// NewTransactionMonitor creates a new instance of TransactionMonitor
func NewTransactionMonitor(ctx context.Context, config Config) (*TransactionMonitor, error) {
	client, err := ethclient.Dial(config.EthereumNode)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}
	credential := options.Credential{
		Username: config.MongoUser,
		Password: config.MongoPassword,
	}

	mongoOptions := options.Client().
		ApplyURI(config.MongoURI).
		SetAuth(credential).
		SetTimeout(10 * time.Second).
		SetRetryWrites(true)

	// Initialize MongoDB client
	mongoClient, err := mongo.Connect(ctx, mongoOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping MongoDB to verify connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	// Rate limiter to handle request limits (1 request per second)
	limiter := rate.NewLimiter(rate.Every(1*time.Second), 1)

	return &TransactionMonitor{
		config:    config,
		client:    client,
		db:        mongoClient.Database(config.DatabaseName),
		txCache:   &sync.Map{},
		alertChan: make(chan Alert, 100),
		limiter:   limiter,
	}, nil
}

// Start begins monitoring Ethereum transactions
func (tm *TransactionMonitor) Start(ctx context.Context) error {
	// Start alert processor
	go tm.processAlerts(ctx)

	query := ethereum.FilterQuery{
		Addresses: append(tm.config.WatchedTokens, tm.config.WatchedContracts...),
	}

	logs := make(chan types.Log)
	sub, err := tm.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			return fmt.Errorf("subscription error: %w", err)
		case vLog := <-logs:
			go tm.processLog(ctx, vLog)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Retry logic with exponential backoff
func (tm *TransactionMonitor) processLog(ctx context.Context, log types.Log) {
	// Check if we've already processed this transaction
	if _, exists := tm.txCache.Load(log.TxHash.Hex()); exists {
		return
	}
	tm.txCache.Store(log.TxHash.Hex(), true)

	// Retry with exponential backoff in case of failure
	retries := 0
	maxRetries := 5

	for {
		err := tm.limiter.Wait(ctx) // Wait for rate limiter
		if err != nil {
			fmt.Printf("Error waiting for rate limiter: %v", err)
			return
		}

		tx, isPending, err := tm.client.TransactionByHash(ctx, log.TxHash)
		if err != nil {
			if retries < maxRetries && (isTooManyRequests(err) || isRateLimitError(err)) {
				retries++
				backoffDuration := time.Duration(retries*2) * time.Second
				fmt.Printf("Retrying transaction fetch due to too many requests, attempt %d, backoff %v", retries, backoffDuration)
				time.Sleep(backoffDuration)
				continue
			}
			fmt.Printf("Error getting transaction: %v", err)
			return
		}
		if isPending {
			fmt.Printf("Transaction pending: %v", log.TxHash.Hex())
			return
		}

		// Process transaction
		tm.handleTransaction(ctx, tx, log.TxHash)
		break
	}
}

func (tm *TransactionMonitor) handleTransaction(ctx context.Context, tx *types.Transaction, txHash common.Hash) {
	// Get transaction sender
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		log.Printf("Error getting transaction sender: %v", err)
		return
	}

	value := new(big.Float).Quo(
		new(big.Float).SetInt(tx.Value()),
		new(big.Float).SetFloat64(1e18),
	)
	valueFloat, _ := value.Float64()

	transaction := Transaction{
		ID:           primitive.NewObjectID(),
		Hash:         tx.Hash().Hex(),
		From:         from.Hex(),
		To:           tx.To().Hex(),
		Value:        valueFloat,
		IsHighValue:  valueFloat >= tm.config.HighValueThreshold,
		IsSuspicious: false,
		Timestamp:    time.Now(),
	}

	if transaction.IsHighValue {
		transaction.AlertType = "HighValue"
		tm.alertChan <- Alert{
			ID:          primitive.NewObjectID(),
			Type:        "HighValue",
			Description: fmt.Sprintf("High value transaction detected: %.2f ETH", valueFloat),
			TxHash:      transaction.Hash,
			Timestamp:   time.Now(),
		}
	}

	if err := tm.saveTransaction(ctx, transaction); err != nil {
		log.Printf("Error saving transaction: %v", err)
	}
}

func isTooManyRequests(err error) bool {
	if err != nil && strings.Contains(err.Error(), "Too Many Requests") {
		return true
	}
	return false
}

func isRateLimitError(err error) bool {
	return err.Error() == "rate limit exceeded"
}

func (tm *TransactionMonitor) saveTransaction(ctx context.Context, tx Transaction) error {
	_, err := tm.db.Collection("transactions").InsertOne(ctx, tx)
	return err
}

func (tm *TransactionMonitor) processAlerts(ctx context.Context) {
	for {
		select {
		case alert := <-tm.alertChan:
			if err := tm.saveAlert(ctx, alert); err != nil {
				log.Printf("Error saving alert: %v", err)
			}
			tm.notifyAlert(alert)
		case <-ctx.Done():
			return
		}
	}
}

func (tm *TransactionMonitor) saveAlert(ctx context.Context, alert Alert) error {
	_, err := tm.db.Collection("alerts").InsertOne(ctx, alert)
	return err
}

func (tm *TransactionMonitor) notifyAlert(alert Alert) {
	// Implement notification logic (e.g., email, Slack, Discord)
	log.Printf("ALERT: %s - %s (TX: %s)", alert.Type, alert.Description, alert.TxHash)
}

func main() {
	ctx := context.Background()

	config := Config{
		EthereumNode:       "wss://mainnet.infura.io/ws/v3/Api-key-here",
		HighValueThreshold: 100,
		WatchedTokens: []common.Address{
			common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), // DAI
			common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606EB48"), // USDC
		},
		WatchedContracts: []common.Address{
			common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
			common.HexToAddress("0x5d3a536E4D6DbD6114cc1Ead35777bAB948E3643"), // Compound cDAI
		},
		MongoURI:      "mongodb://localhost:27017",
		MongoUser:     "root",
		MongoPassword: "example",
		DatabaseName:  "ethereum_monitor",
	}

	monitor, err := NewTransactionMonitor(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create transaction monitor: %v", err)
	}

	log.Println("Starting Ethereum transaction monitor...")
	if err := monitor.Start(ctx); err != nil {
		log.Fatalf("Monitor failed: %v", err)
	}
}
