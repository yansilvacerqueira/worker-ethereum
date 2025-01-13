# Ethereum Transaction Monitor

OBS: This is a study project, and there are still several improvements to be made. Feel free to contribute with PRs and new ideas.

### Overview

The Ethereum Transaction Monitor is a robust Go application designed to monitor Ethereum blockchain transactions in real-time. It tracks specific token and contract addresses, detects high-value transactions, and generates alerts based on configurable thresholds. The application consists of two main components: a transaction monitoring worker and a REST API server.

### Features

- Real-time monitoring of Ethereum transactions
- Tracking of specific ERC-20 tokens and smart contracts
- High-value transaction detection and alerting
- Transaction and alert data storage in MongoDB
- REST API endpoints for accessing historical data
- Rate limiting and retry mechanisms for robust operation
- Graceful shutdown handling

### Prerequisites

- Go 1.19 or higher
- Docker and Docker Compose
- MongoDB
- Ethereum Node access (e.g., Infura WebSocket endpoint)

### Configuration

The application requires the following configuration:

- Ethereum Node WebSocket URL
- MongoDB connection details
- Watched token and contract addresses
- High-value transaction threshold
- API server port (default: 8080)

### Quick Start

1. Clone the repository:

```bash
git clone https://github.com/yansilvacerqueira/worker-ethereum
cd worker-ethereum
```

2. Start MongoDB using Docker Compose:

```bash
docker-compose up -d
```

3. Set your environment variables:

```bash
export ETHEREUM_NODE="wss://mainnet.infura.io/ws/v3/your-api-key"
export MONGO_URI="mongodb://localhost:27017"
export MONGO_USER="root"
export MONGO_PASSWORD="example"
```

4. Run the application:

```bash
go run cmd/main.go
```

### API Endpoints

- `GET /api/transactions` - Retrieve monitored transactions
- `GET /api/alerts` - Retrieve generated alerts

### Docker Compose

The project includes a `docker-compose.yml` file for easy deployment of MongoDB:

```yaml
version: "3.8"
services:
  mongodb:
    image: mongo:latest
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: example
```

### Script to populate the mongo

Run this script to access the container and populate the db with some
test data.

```bash
docker exec -it mongodb mongosh -u root -p example

use ethereum_monitor

db.transactions.insertOne({
  hash: "0x123",
  from: "0xabc",
  to: "0xdef",
  value: 1.5,
  timestamp: new Date()
})

db.alerts.insertOne({
  txHash: "0x123",
  alertType: "high_value",
  description: "Test alert",
  timestamp: new Date()
})
```
