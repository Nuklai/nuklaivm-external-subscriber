// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/types"
	ctypes "github.com/cometbft/cometbft/types"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Wrapper for the tendermint RPC client
type Client struct {
	clients []*comethttp.HTTP
	mutex   sync.RWMutex
}

// New RPC client
func NewClient(endpoints []string) (*Client, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("no endpoints provided")
	}

	clients := make([]*comethttp.HTTP, 0, len(endpoints))
	for _, endpoint := range endpoints {
		client, err := comethttp.New(endpoint)
		if err != nil {
			logger.Warn("Failed to create RPC client", zap.String("endpoint", endpoint), zap.Error(err))
			continue
		}
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return nil, errors.New("no valid endpoints provided")
	}

	logger.Info("Created RPC clients", zap.Int("count", len(clients)), zap.Strings("endpoints", endpoints))
	return &Client{
		clients: clients,
	}, nil
}

// Start all clients
func (c *Client) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, client := range c.clients {
		if err := client.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop all clients
func (c *Client) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, client := range c.clients {
		if err := client.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// Returns a random healthy RPC client
func (c *Client) getClient() (*comethttp.HTTP, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.clients) == 0 {
		return nil, errors.New("no RPC clients available")
	}

	// Use a more secure random number
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(len(c.clients))
	return c.clients[idx], nil
}

// Returns the node status
func (c *Client) Status(ctx context.Context) (*ctypes.ResultStatus, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	status, err := client.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %w", err)
	}

	return status, nil
}

// Returns the block at the specified height
func (c *Client) Block(ctx context.Context, height int64) (*ctypes.ResultBlock, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	block, err := client.Block(ctx, &height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	return block, nil
}

// Returns the block results at the specified height
func (c *Client) BlockResults(ctx context.Context, height int64) (*ctypes.ResultBlockResults, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	results, err := client.BlockResults(ctx, &height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block results at height %d: %w", height, err)
	}

	return results, nil
}

// Returns the validators at the specified height
func (c *Client) Validators(ctx context.Context, height int64, page, perPage int) (*ctypes.ResultValidators, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	vals, err := client.Validators(ctx, &height, &page, &perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators at height %d: %w", height, err)
	}

	return vals, nil
}

// Tx returns the transaction with the specified hash
func (c *Client) Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	tx, err := client.Tx(ctx, hash, prove)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

// Start's a search for transactions by query
func (c *Client) TxSearch(ctx context.Context, query string, prove bool, page, perPage int, orderBy string) (*ctypes.ResultTxSearch, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	results, err := client.TxSearch(ctx, query, prove, &page, &perPage, orderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to search transactions: %w", err)
	}

	return results, nil
}

// Performs an ABCI query
func (c *Client) ABCIQuery(ctx context.Context, path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	result, err := client.ABCIQuery(ctx, path, data)
	if err != nil {
		return nil, fmt.Errorf("failed to perform ABCI query: %w", err)
	}

	return result, nil
}

// Broadcasts a transaction synchronously
func (c *Client) BroadcastTxSync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	result, err := client.BroadcastTxSync(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction synchronously: %w", err)
	}

	return result, nil
}

// Returns the latest block height
func (c *Client) LatestHeight(ctx context.Context) (int64, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

// Health
func (c *Client) IsHealthy(ctx context.Context) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, client := range c.clients {
		_, err := client.Status(ctx)
		if err == nil {
			return true
		}
	}

	return false
}
