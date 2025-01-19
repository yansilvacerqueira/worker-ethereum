package rpc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Manager handles multiple RPC connections with round-robin load balancing
type Manager struct {
	nodes     []string
	clients   []*ethclient.Client
	current   uint32
	connected []bool
	mu        sync.RWMutex
}

// NewManager creates a new RPC manager with the given nodes
func NewManager(nodes []string) (*Manager, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no RPC nodes provided")
	}

	manager := &Manager{
		nodes:     nodes,
		clients:   make([]*ethclient.Client, len(nodes)),
		connected: make([]bool, len(nodes)),
	}

	// Inicializa as conexões
	for i, node := range nodes {
		client, err := ethclient.Dial(node)
		if err != nil {
			log.Printf("Failed to connect to node %s: %v", node, err)
			continue
		}

		manager.clients[i] = client
		manager.connected[i] = true
	}

	// Verifica se pelo menos um nó está conectado
	if !manager.hasConnectedNodes() {
		return nil, fmt.Errorf("failed to connect to any RPC nodes")
	}

	// Inicia o health check
	go manager.startHealthCheck()

	return manager, nil
}

func (rm *Manager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)

	for range ticker.C {
		rm.checkNodes()
	}
}

func (rm *Manager) hasConnectedNodes() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for _, connected := range rm.connected {
		if connected {
			return true
		}
	}
	return false
}

func (rm *Manager) GetClient() *ethclient.Client {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Implementação do round-robin
	current := atomic.AddUint32(&rm.current, 1)
	for i := 0; i < len(rm.clients); i++ {
		idx := (int(current) + i) % len(rm.clients)

		if rm.connected[idx] && rm.clients[idx] != nil {
			return rm.clients[idx]
		}
	}

	return nil
}

func (rm *Manager) checkNodes() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for i, client := range rm.clients {
		if client == nil {
			// Tenta reconectar nós desconectados
			if newClient, err := ethclient.Dial(rm.nodes[i]); err == nil {
				rm.clients[i] = newClient
				rm.connected[i] = true
				log.Printf("Reconnected to node %s", rm.nodes[i])
			}
			continue
		}

		// Verifica se o nó está respondendo
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.BlockNumber(ctx)
		cancel()

		if err != nil {
			log.Printf("Node %s is not responding: %v", rm.nodes[i], err)
			rm.connected[i] = false
			client.Close()
			rm.clients[i] = nil
		} else {
			rm.connected[i] = true
		}
	}
}
