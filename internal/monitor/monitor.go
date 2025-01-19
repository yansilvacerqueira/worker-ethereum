package monitor

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/yansilvacerqueira/worker-ethereum/internal/config"
	"github.com/yansilvacerqueira/worker-ethereum/internal/models"

	"github.com/yansilvacerqueira/worker-ethereum/internal/rpc"

	"github.com/yansilvacerqueira/worker-ethereum/internal/mongodb"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/time/rate"
)

// Monitor handles the transaction monitoring process
type Monitor struct {
	config     *config.Config
	rpcManager *rpc.Manager
	storage    *mongodb.Client
	alertChan  chan models.Alert
	txCache    *sync.Map
	limiter    *rate.Limiter
}

// New creates a new transaction monitor
func New(ctx context.Context, cfg *config.Config) (*Monitor, error) {
	rpcManager, err := rpc.NewManager(cfg.EthereumNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC manager: %w", err)
	}

	db, err := mongodb.NewClient(ctx, cfg.MongoURI, cfg.MongoUser, cfg.MongoPassword, cfg.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return &Monitor{
		config:     cfg,
		rpcManager: rpcManager,
		storage:    db,
		alertChan:  make(chan models.Alert, 100),
		txCache:    &sync.Map{},
		limiter:    rate.NewLimiter(rate.Every(1*time.Second), 1),
	}, nil
}

// Start begins monitoring Ethereum transactions
func (tm *Monitor) Start(ctx context.Context) error {
	go tm.processAlerts(ctx)

	for {
		client := tm.rpcManager.GetClient()
		if client == nil {
			log.Println("No RPC clients available, waiting...")
			time.Sleep(5 * time.Second)
			continue
		}

		query := ethereum.FilterQuery{
			Addresses: append(tm.config.WatchedTokens, tm.config.WatchedContracts...),
		}

		logs := make(chan types.Log)
		sub, err := client.SubscribeFilterLogs(ctx, query, logs)
		if err != nil {
			log.Printf("Failed to subscribe to logs: %v, trying another client...", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for {
			select {
			case err := <-sub.Err():
				log.Printf("Subscription error: %v, reconnecting...", err)
				sub.Unsubscribe()
				time.Sleep(1 * time.Second)
			case vLog := <-logs:
				go tm.processLog(ctx, vLog, client)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (tm *Monitor) processAlerts(ctx context.Context) {
	mongoClient := tm.storage
	for {
		select {
		case alert := <-tm.alertChan:
			if err := mongoClient.SaveAlert(ctx, alert); err != nil {
				log.Printf("Error saving alert: %v", err)
			}
			tm.notifyAlert(alert)
		case <-ctx.Done():
			return
		}
	}
}

func (tm *Monitor) notifyAlert(alert models.Alert) {
	// Implement notification logic (e.g., email, Slack, Discord)
	log.Printf("ALERT: %s - %s (TX: %s)", alert.AlertType, alert.Description, alert.TxHash)
}

// Retry logic with exponential backoff
func (tm *Monitor) processLog(ctx context.Context, log types.Log, client *ethclient.Client) {
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

		tx, isPending, err := client.TransactionByHash(ctx, log.TxHash)
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
		tm.handleTransaction(ctx, tx)
		break
	}
}

func (tm *Monitor) handleTransaction(ctx context.Context, tx *types.Transaction) {
	mongoClient := tm.storage
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

	transaction := models.Transaction{
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
		tm.alertChan <- models.Alert{
			ID:          primitive.NewObjectID(),
			AlertType:   "HighValue",
			Description: fmt.Sprintf("High value transaction detected: %.2f ETH", valueFloat),
			TxHash:      transaction.Hash,
			Timestamp:   time.Now(),
		}
	}

	if err := mongoClient.SaveTransaction(ctx, transaction); err != nil {
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
