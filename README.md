# Ethereum Transaction Monitor

A robust Ethereum transaction monitoring system that tracks high-value transactions and specific token movements on the Ethereum blockchain. Built with Go, it features automatic failover between multiple RPC nodes, MongoDB storage, and a REST API for data access.

## Features

- 🔍 Real-time monitoring of Ethereum transactions
- 💰 High-value transaction detection and alerting
- 🔄 Multiple RPC node support with automatic failover
- 📊 REST API for querying transaction history and alerts
- 🔐 MongoDB storage for transaction and alert data
- ⚡ Rate limiting and retry mechanisms
- 🛡️ Graceful shutdown handling
- 🐳 Docker support with MongoDB and mongo-express

## Architecture

The project consists of two main components:

1. **Worker Service**: Monitors the Ethereum blockchain for transactions
2. **API Service**: Provides HTTP endpoints to query stored data

### Key Components:

- `monitor`: Core transaction monitoring logic
- `rpc`: RPC connection management with failover
- `mongodb`: Database operations
- `config`: Application configuration
- `models`: Data structures
- `api`: REST API handlers

## Installation

### Standard Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/ethereum-monitor
cd ethereum-monitor

# Install dependencies
go mod download

# Build the project
go build -o worker ./cmd/worker
go build -o api ./cmd/api
```

### Docker Installation

The project includes a `docker-compose.yml` file for easy deployment of the MongoDB database and mongo-express interface.

```bash
# Start MongoDB and mongo-express
docker-compose up -d
```

This will start:

- MongoDB on port 27017
- Mongo Express (web interface) on port 8081

#### MongoDB Configuration:

- Username: root
- Password: example
- Database: ethereum_monitor

#### Mongo Express Configuration:

- URL: http://localhost:8081
- Username: admin
- Password: pass

#### Docker Volumes:

- `mongodb_data`: Persistent storage for MongoDB data

## Configuration

The system uses a configuration structure defined in `config.go`. You can customize the following parameters:

```go
type Config struct {
  EthereumNodes      []string         // RPC node URLs
  HighValueThreshold float64          // Threshold for high-value alerts
  WatchedTokens      []common.Address // Token contracts to monitor
  WatchedContracts   []common.Address // Other contracts to monitor
  MongoURI           string           // MongoDB connection URI
  MongoUser          string           // MongoDB username
  MongoPassword      string           // MongoDB password
  DatabaseName       string           // MongoDB database name
}
```

### Default Configuration

```go
cfg := config.NewDefaultConfig()
// Includes:
// - Multiple public Ethereum RPC nodes
// - 100 ETH high-value threshold
// - Common tokens (DAI, USDC, WETH, cDAI)
// - Local MongoDB connection
```

## Usage

### Starting the Worker

```bash
./worker
```

The worker will:

1. Connect to configured RPC nodes
2. Monitor transactions for watched addresses
3. Generate alerts for high-value transactions
4. Store transaction data in MongoDB

### Starting the API Server

```bash
./api
```

The API server provides the following endpoints:

- `GET /api/transactions`: Retrieve stored transactions
- `GET /api/alerts`: Retrieve generated alerts

## API Examples

### Get Transactions

```bash
curl http://localhost:8080/api/transactions
```

### Get Alerts

```bash
curl http://localhost:8080/api/alerts
```

## RPC Failover System

The system includes a sophisticated RPC management system that:

1. Maintains connections to multiple Ethereum nodes
2. Performs health checks every 30 seconds
3. Automatically switches to healthy nodes when failures occur
4. Implements round-robin load balancing

Example configuration of RPC nodes:

```go
EthereumNodes: []string{
  "wss://ethereum.publicnode.com",
  "wss://mainnet.gateway.tenderly.co",
  "wss://rpc.ankr.com/eth/ws",
}
```

## MongoDB Schema

### Transaction Collection

```json
{
  "_id": ObjectId,
  "hash": String,
  "from": String,
  "to": String,
  "value": Number,
  "token_name": String,
  "token_value": Number,
  "is_high_value": Boolean,
  "is_suspicious": Boolean,
  "alert_type": String,
  "timestamp": Date
}
```

### Alert Collection

```json
{
  "_id": ObjectId,
  "tx_hash": String,
  "alert_type": String,
  "description": String,
  "timestamp": Date
}
```

## Error Handling

The system implements robust error handling:

- Exponential backoff for RPC request retries
- Rate limiting to prevent API abuse
- Graceful shutdown handling
- Comprehensive error logging

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
