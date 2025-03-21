// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package core

import (
	"context"
	"fmt"
	"time"

	"github.com/nuklai/nuklaivm-external-subscriber/client/rpc"
	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"github.com/nuklai/nuklaivm-external-subscriber/core/blocks"
	"github.com/nuklai/nuklaivm-external-subscriber/db/postgres"
	"github.com/nuklai/nuklaivm-external-subscriber/db/repository"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Indexer core
type Indexer struct {
	cfg  *config.Config
	db   *postgres.Database
	repo repository.Repository

	rpcClient *rpc.Client

	blockProcessor *blocks.Processor
	blockSyncer    *blocks.Syncer
	blockFetcher   *blocks.Fetcher
	stopCh         chan struct{}
}

func NewIndexer(cfg *config.Config) (*Indexer, error) {
	db, err := postgres.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create repo
	repo := repository.NewPostgresRepository(db)

	// Create RPC client
	rpcClient, err := rpc.NewClient(cfg.Blockchain.RPCEndpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	// Create block fetcher
	blockFetcher := blocks.NewFetcher(rpcClient)

	// Create block processor
	blockProcessor := blocks.NewProcessor(blockFetcher, repo, rpcClient)

	// Create block syncer
	blockSyncer := blocks.NewSyncer(blockProcessor, &cfg.Blockchain)

	return &Indexer{
		cfg:            cfg,
		db:             db,
		repo:           repo,
		rpcClient:      rpcClient,
		blockProcessor: blockProcessor,
		blockSyncer:    blockSyncer,
		blockFetcher:   blockFetcher,
		stopCh:         make(chan struct{}),
	}, nil
}

// Start the indexer
func (i *Indexer) Start(ctx context.Context) error {
	logger.Info("Starting indexer...")

	if err := i.db.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if !i.rpcClient.IsHealthy(ctx) {
		return fmt.Errorf("RPC client is not healthy")
	}

	if err := i.blockSyncer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start block syncer: %w", err)
	}

	logger.Info("Indexer started successfully")

	go i.monitorSyncStatus(ctx)

	return nil
}

// Stop the indexer
func (i *Indexer) Stop() {
	logger.Info("Stopping indexer...")

	i.blockSyncer.Stop()

	i.db.Close()

	close(i.stopCh)

	logger.Info("Indexer stopped")
}

func (i *Indexer) monitorSyncStatus(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-i.stopCh:
			return
		case <-ticker.C:
			synced, latest, running := i.blockSyncer.GetSyncStatus()

			var syncPercentage float64
			if latest > 0 {
				syncPercentage = float64(synced) / float64(latest) * 100
			}

			logger.Info("Sync status",
				zap.Int64("synced_height", synced),
				zap.Int64("latest_height", latest),
				zap.Float64("sync_percentage", syncPercentage),
				zap.Bool("syncer_running", running),
			)
		}
	}
}

// GetRepository returns the indexer's repository
func (i *Indexer) GetRepository() repository.Repository {
	return i.repo
}
