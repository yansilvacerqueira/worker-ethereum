package ethereum

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	client *ethclient.Client
}

func NewClient(nodeURL string) (*Client, error) {
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

func (c *Client) SubscribeToNewBlocks(ctx context.Context) (<-chan *types.Header, error) {
	headers := make(chan *types.Header)
	sub, err := c.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				log.Printf("subscription error: %v", err)
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return headers, nil
}
