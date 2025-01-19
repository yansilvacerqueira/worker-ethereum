**Transações Ethereum**

Sistema de monitoramento de transações Ethereum que rastreia transações de alto valor e movimentos específicos de tokens na blockchain Ethereum. Construído com Go, ele apresenta failover automático entre múltiplos nós RPC, armazenamento em MongoDB e uma API REST para acesso aos dados.

## Funcionalidades

- 🔍 **Monitoramento em tempo real** das transações Ethereum
- 💰 **Detecção e alerta** de transações de alto valor
- 🔄 **Suporte a múltiplos nós RPC** com failover automático
- 📊 **API REST** para consulta do histórico de transações e alertas
- 🔐 **Armazenamento em MongoDB** para dados de transação e alerta
- ⚡ **Limitação de taxa e mecanismos de repetição**
- 🛡️ **Tratamento de desligamento suave**
- 🐳 **Suporte a Docker** com MongoDB e mongo-express

## Arquitetura

O projeto consiste em dois componentes principais:

1. **Serviço Worker**: Monitora a blockchain Ethereum para transações
2. **Serviço API**: Fornece endpoints HTTP para consultar dados armazenados

### Componentes Principais:

- `monitor`: Lógica central de monitoramento de transações
- `rpc`: Gerenciamento de conexão RPC com failover
- `mongodb`: Operações no banco de dados
- `config`: Configuração da aplicação
- `models`: Estruturas de dados
- `api`: Manipuladores da API REST

## Instalação

### Instalação Padrão

```bash
# Clone o repositório
git clone https://github.com/yourusername/ethereum-monitor
cd ethereum-monitor

# Instale as dependências
go mod download

# Construa o projeto
go build -o worker ./cmd/worker
go build -o api ./cmd/api
```

### Instalação via Docker

O projeto inclui um arquivo `docker-compose.yml` para fácil implantação do banco de dados MongoDB e da interface mongo-express.

```bash
# Inicie o MongoDB e mongo-express
docker-compose up -d
```

Isso iniciará:

- MongoDB na porta 27017
- Mongo Express (interface web) na porta 8081

#### Configuração do MongoDB:

- Nome de usuário: root
- Senha: example
- Banco de dados: ethereum_monitor

#### Configuração do Mongo Express:

- URL: http://localhost:8081
- Nome de usuário: admin
- Senha: pass

#### Volumes do Docker:

- `mongodb_data`: Armazenamento persistente para dados do MongoDB

## Configuração

O sistema utiliza uma estrutura de configuração definida em `config.go`. Você pode personalizar os seguintes parâmetros:

```go
type Config struct {
    EthereumNodes []string // URLs dos nós RPC
    HighValueThreshold float64 // Limite para alertas de alto valor
    WatchedTokens []common.Address // Contratos de tokens a serem monitorados
    WatchedContracts []common.Address // Outros contratos a serem monitorados
    MongoURI string // URI de conexão com o MongoDB
    MongoUser string // Nome de usuário do MongoDB
    MongoPassword string // Senha do MongoDB
    DatabaseName string // Nome do banco de dados do MongoDB
}
```

### Configuração Padrão

```go
cfg := config.NewDefaultConfig()
// Inclui:
// - Múltiplos nós RPC públicos da Ethereum
// - Limite alto de 100 ETH
// - Tokens comuns (DAI, USDC, WETH, cDAI)
// - Conexão local com o MongoDB
```

## Uso

### Iniciando o Worker

```bash
./worker
```

O worker irá:

1. Conectar-se aos nós RPC configurados.
2. Monitorar transações para endereços observados.
3. Gerar alertas para transações de alto valor.
4. Armazenar dados das transações no MongoDB.

### Iniciando o Servidor API

```bash
./api
```

O servidor API fornece os seguintes endpoints:

- `GET /api/transactions`: Recuperar transações armazenadas.
- `GET /api/alerts`: Recuperar alertas gerados.

## Exemplos da API

### Obter Transações

```bash
curl http://localhost:8080/api/transactions
```

### Obter Alertas

```bash
curl http://localhost:8080/api/alerts
```

## Sistema de Failover RPC

O sistema inclui um "sofisticado" sistema de gerenciamento RPC que:

1. Mantém conexões com múltiplos nós Ethereum.
2. Realiza verificações de saúde a cada 30 segundos.
3. Alterna automaticamente para nós saudáveis quando falhas ocorrem.
4. Implementa balanceamento de carga round-robin.

Exemplo de configuração dos nós RPC:

```go
EthereumNodes: []string{
    "wss://ethereum.publicnode.com",
    "wss://mainnet.gateway.tenderly.co",
    "wss://rpc.ankr.com/eth/ws",
}
```

## Esquema do MongoDB

### Coleção de Transações

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

### Coleção de Alertas

```json
{
    "_id": ObjectId,
    "tx_hash": String,
    "alert_type": String,
    "description": String,
    "timestamp": Date
}
```

## Tratamento de Erros

Tratamento de erros:

- Repetições com backoff exponencial para requisições RPC.
- Limitação da taxa para prevenir abusos da API.
- Tratamento suave no desligamento.
- Registro abrangente dos erros.

## Deseja Contribuir ?

1. Faça um fork do repositório.
2. Crie sua branch de recurso (`git checkout -b feature/amazing-feature`).
3. Comite suas alterações (`git commit -m 'Adicione uma funcionalidade incrível'`).
4. Envie para a branch (`git push origin feature/amazing-feature`).
5. Abra um Pull Request.
