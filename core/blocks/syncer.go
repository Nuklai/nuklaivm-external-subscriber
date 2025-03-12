// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"context"
	"fmt"
	"github.com/nuklai/nuklaivm-external-subscriber/db/repository"
	"sync"
	"time"

	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Responsible for syncing blocks
type Syncer struct {
	processor       *Processor
	startHeight     int64
	syncConcurrency int
	fetchInterval   time.Duration
	fetchTimeout    time.Duration
	running         bool
	stopCh          chan struct{}
	mutex           sync.Mutex

	// Track sync progress
	latestSyncedHeight int64
	highestKnownHeight int64
}

func NewSyncer(processor *Processor, cfg *config.BlockchainConfig) *Syncer {
	return &Syncer{
		processor:       processor,
		startHeight:     cfg.StartHeight,
		syncConcurrency: cfg.SyncConcurrency,
		fetchInterval:   cfg.FetchInterval,
		fetchTimeout:    cfg.FetchTimeout,
		running:         false,
		stopCh:          make(chan struct{}),
	}
}

// Start starts the block syncer
func (s *Syncer) Start(ctx context.Context) error {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return fmt.Errorf("syncer is already running")
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mutex.Unlock()

	// goroutine for historical sync
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("Context cancelled, stopping historical sync")
				return
			case <-s.stopCh:
				logger.Info("Received stop signal, stopping historical sync")
				return
			default:
				err := s.syncHistoricalBlocks(ctx)
				if err != nil {
					logger.Error("Failed to sync historical blocks", zap.Error(err))
				}

				time.Sleep(s.fetchInterval)
			}
		}
	}()

	// goroutine for monitoring latest blocks
	go func() {
		// Ticker
		ticker := time.NewTicker(s.fetchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Context cancelled, stopping latest block monitoring")
				return
			case <-s.stopCh:
				logger.Info("Received stop signal, stopping latest block monitoring")
				return
			case <-ticker.C:
				err := s.updateHighestKnownHeight(ctx)
				if err != nil {
					logger.Error("Failed to update highest known height", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// Stop Syncer
func (s *Syncer) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	close(s.stopCh)
	s.running = false

	logger.Info("Block syncer stopped")
}

// Sync historical blocks from the chain
func (s *Syncer) syncHistoricalBlocks(ctx context.Context) error {
	err := s.updateHighestKnownHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to update highest known height: %w", err)
	}

	startHeight, endHeight, err := s.getSyncRange(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sync range: %w", err)
	}

	if startHeight > endHeight {
		logger.Info("Historical sync is caught up",
			zap.Int64("latest_synced", s.latestSyncedHeight),
			zap.Int64("highest_known", s.highestKnownHeight),
		)
		return nil
	}

	logger.Info("Starting historical sync",
		zap.Int64("start_height", startHeight),
		zap.Int64("end_height", endHeight),
		zap.Int("concurrency", s.syncConcurrency),
	)

	// Wworker pool for processing the blocks
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.syncConcurrency)
	errorCh := make(chan error, s.syncConcurrency)

	// Track heights
	successHeights := make(chan int64, s.syncConcurrency)

	// goroutine to update latestSyncedHeight
	go func() {
		for height := range successHeights {
			s.mutex.Lock()
			if height > s.latestSyncedHeight {
				s.latestSyncedHeight = height
			}
			s.mutex.Unlock()
		}
	}()

	for height := startHeight; height <= endHeight; height++ {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, stopping block processing")
			wg.Wait()
			close(successHeights)
			return ctx.Err()
		case <-s.stopCh:
			logger.Info("Received stop signal, stopping block processing")
			wg.Wait()
			close(successHeights)
			return nil
		case err := <-errorCh:
			logger.Error("Error encountered during block processing", zap.Error(err))
			wg.Wait()
			close(successHeights)
			return err
		default:
			// Continue processing the block
		}

		semaphore <- struct{}{}

		wg.Add(1)
		go func(h int64) {
			defer wg.Done()
			defer func() { <-semaphore }()

			blockCtx, cancel := context.WithTimeout(ctx, s.fetchTimeout)
			defer cancel()

			err := s.processor.ProcessBlock(blockCtx, h)
			if err != nil {
				errorCh <- fmt.Errorf("failed to process block at height %d: %w", h, err)
				return
			}

			successHeights <- h
		}(height)
	}

	wg.Wait()
	close(successHeights)

	logger.Info("Historical sync completed",
		zap.Int64("start_height", startHeight),
		zap.Int64("end_height", endHeight),
		zap.Int64("latest_synced", s.latestSyncedHeight),
	)

	return nil
}

func (s *Syncer) updateHighestKnownHeight(ctx context.Context) error {
	height, err := s.processor.rpcClient.LatestHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest height: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if height > s.highestKnownHeight {
		logger.Debug("Updated highest known height",
			zap.Int64("old", s.highestKnownHeight),
			zap.Int64("new", height),
		)
		s.highestKnownHeight = height
	}

	return nil
}

// Returns the start and end heights for the next sync batch
func (s *Syncer) getSyncRange(ctx context.Context) (int64, int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.latestSyncedHeight == 0 {
		latestBlock, err := s.processor.repository.GetLatestBlock(ctx)
		if err != nil {
			if err == repository.ErrNotFound {
				s.latestSyncedHeight = s.startHeight - 1
			} else {
				return 0, 0, fmt.Errorf("failed to get latest block from repository: %w", err)
			}
		} else {
			s.latestSyncedHeight = latestBlock.Height
		}
	}

	// Calculate the start and end heights for next batch
	startHeight := s.latestSyncedHeight + 1
	endHeight := s.highestKnownHeight

	batchSize := int64(s.syncConcurrency * 10) // Heuristic: 10x concurrency
	if endHeight-startHeight+1 > batchSize {
		endHeight = startHeight + batchSize - 1
	}

	return startHeight, endHeight, nil
}

// Returns sync status
func (s *Syncer) GetSyncStatus() (int64, int64, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.latestSyncedHeight, s.highestKnownHeight, s.running
}
