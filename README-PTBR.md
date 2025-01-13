# Transações Ethereum

OBS: Este é um projeto de estudo e ainda há diversas melhorias a serem feitas. Fique à vontade para contribuir com PRs e novas ideias.

### Visão Geral

A aplicação em Go projetada para acompanhar transações na blockchain Ethereum em tempo real. Ele monitora endereços de tokens e contratos específicos, identifica transações de alto valor e emite alertas com base em limites configuráveis. A aplicação é composta por dois principais componentes: um worker para monitorar as transações e um servidor de API REST.

### Funcionalidades

- Monitoramento em tempo real de transações na rede Ethereum
- Rastreamento de tokens ERC-20 e contratos inteligentes específicos
- Detecção e alertas para transações de alto valor
- Armazenamento de dados de transações e alertas no MongoDB
- Endpoints REST API para consultar dados históricos
- Mecanismos de limitação de taxa e tentativas automáticas para garantir estabilidade
- Gerenciamento de encerramento de forma segura

### Pré-requisitos

- Go 1.19 ou superior
- Docker e Docker Compose
- MongoDB
- Acesso a um nó Ethereum (ex: WebSocket via Infura)

### Configuração

A aplicação precisa das seguintes configurações:

- URL WebSocket do nó Ethereum
- Informações de conexão com o MongoDB
- Endereços de tokens e contratos a serem monitorados
- Limite para alertar sobre transações de alto valor
- Porta para o servidor API (padrão: 8080)

### Início Rápido

1. Clone o repositório:

```bash
git clone https://github.com/yansilvacerqueira/worker-ethereum
cd worker-ethereum
```

2. Inicie o MongoDB com Docker Compose:

```bash
docker-compose up -d
```

3. Configure suas variáveis de ambiente:

```bash
export ETHEREUM_NODE="wss://mainnet.infura.io/ws/v3/sua-api-key"
export MONGO_URI="mongodb://localhost:27017"
export MONGO_USER="root"
export MONGO_PASSWORD="example"
```

4. Execute a aplicação:

```bash
go run cmd/main.go
```

### Endpoints da API

- `GET /api/transactions` - Retorna as transações monitoradas
- `GET /api/alerts` - Retorna os alertas gerados

### Docker Compose

O projeto inclui um arquivo `docker-compose.yml` para facilitar a configuração do MongoDB:

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

### Script para popular o mongo

Execute este script para acessar o container e popular o db com alguns
dados de teste.

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
