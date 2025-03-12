// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
	"github.com/nuklai/nuklaivm-external-subscriber/client/rpc"
	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
	"github.com/nuklai/nuklaivm-external-subscriber/db/repository"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Processor is in charge of processing blocks
type Processor struct {
	rpcClient  *rpc.Client
	repository repository.Repository
}

func NewProcessor(rpcClient *rpc.Client, repo repository.Repository) *Processor {
	return &Processor{
		rpcClient:  rpcClient,
		repository: repo,
	}
}

// Processes a block at the specified height
func (p *Processor) ProcessBlock(ctx context.Context, height int64) error {
	// Get the block from the RPC
	blockResult, err := p.rpcClient.Block(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	// Get the block results from the RPC
	blockResultsResult, err := p.rpcClient.BlockResults(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to get block results at height %d: %w", height, err)
	}

	// Process and save block
	err = p.saveBlock(ctx, blockResult, blockResultsResult)
	if err != nil {
		return fmt.Errorf("failed to save block at height %d: %w", height, err)
	}

	// Process and save transactions
	err = p.saveTransactions(ctx, blockResult, blockResultsResult)
	if err != nil {
		return fmt.Errorf("failed to save transactions at height %d: %w", height, err)
	}

	// Process begin-block events
	err = p.processBeginBlockEvents(ctx, height, blockResult.Block.Time, blockResultsResult.BeginBlockEvents)
	if err != nil {
		return fmt.Errorf("failed to process begin-block events at height %d: %w", height, err)
	}

	// Process end-block events
	err = p.processEndBlockEvents(ctx, height, blockResult.Block.Time, blockResultsResult.FinalizeBlockEvents)
	if err != nil {
		return fmt.Errorf("failed to process end-block events at height %d: %w", height, err)
	}

	logger.Info("Processed block",
		zap.Int64("height", height),
		zap.Time("time", blockResult.Block.Time),
		zap.Int("num_txs", len(blockResult.Block.Txs)),
	)

	return nil
}

// Save block to the repo
func (p *Processor) saveBlock(ctx context.Context, blockResult *ctypes.ResultBlock, blockResultsResult *ctypes.ResultBlockResults) error {
	block := blockResult.Block

	// Calculate the total gas
	totalGas := int64(0)
	for _, txResult := range blockResultsResult.TxResults {
		totalGas += txResult.GasUsed
	}

	// Create block model
	blockModel := &models.Block{
		Height:             block.Height,
		Hash:               hex.EncodeToString(block.Hash()),
		ProposerAddress:    hex.EncodeToString(block.ProposerAddress),
		NumTxs:             int64(len(block.Txs)),
		TotalGas:           totalGas,
		Size:               int64(block.Size()),
		Time:               block.Time,
		AppHash:            hex.EncodeToString(block.AppHash),
		Version:            int64(block.Version.Block),
		ChainID:            block.ChainID,
		LastBlockID:        block.LastBlockID.String(),
		LastCommitHash:     hex.EncodeToString(block.LastCommitHash),
		ValidatorsHash:     hex.EncodeToString(block.ValidatorsHash),
		NextValidatorsHash: hex.EncodeToString(block.NextValidatorsHash),
		ConsensusHash:      hex.EncodeToString(block.ConsensusHash),
		DataHash:           hex.EncodeToString(block.DataHash),
		EvidenceHash:       hex.EncodeToString(block.EvidenceHash),
		CreatedAt:          time.Now().UTC(),
	}

	// Save to repo
	return p.repository.SaveBlock(ctx, blockModel)
}

// Saves the txn to the repo
func (p *Processor) saveTransactions(ctx context.Context, blockResult *ctypes.ResultBlock, blockResultsResult *ctypes.ResultBlockResults) error {
	block := blockResult.Block
	txResults := blockResultsResult.TxResults

	// Process each transaction
	for i, txBytes := range block.Txs {
		txHash := hex.EncodeToString(txBytes.Hash())
		txResult := txResults[i]

		// Get transaction details
		txDetail, err := p.rpcClient.Tx(ctx, txBytes.Hash(), false)
		if err != nil {
			logger.Warn("Failed to get transaction details",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			continue
		}

		// Extract signers from transaction
		signers, err := extractSigners(txDetail.Tx)
		if err != nil {
			logger.Warn("Failed to extract signers from transaction",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			signers = []string{}
		}

		// Extract messages from transaction
		messages, err := extractMessages(txDetail.Tx)
		if err != nil {
			logger.Warn("Failed to extract messages from transaction",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			messages = models.Messages{}
		}

		// Create txn model
		txModel := &models.Transaction{
			Hash:        txHash,
			BlockHeight: block.Height,
			BlockTime:   block.Time,
			Index:       int32(i),
			GasWanted:   txResult.GasWanted,
			GasUsed:     txResult.GasUsed,
			Fee:         extractFee(txDetail.Tx),
			Memo:        extractMemo(txDetail.Tx),
			ErrorLog:    txResult.Log,
			Status:      getTxStatus(txResult.Code),
			Codespace:   txResult.Codespace,
			Messages:    messages,
			Signers:     models.Signers(signers),
			CreatedAt:   time.Now().UTC(),
		}

		// Save
		err = p.repository.SaveTransaction(ctx, txModel)
		if err != nil {
			return fmt.Errorf("failed to save transaction %s: %w", txHash, err)
		}

		// Process txn events
		err = p.processTxEvents(ctx, block.Height, block.Time, txHash, i, txResult.Events)
		if err != nil {
			return fmt.Errorf("failed to process events for transaction %s: %w", txHash, err)
		}
	}

	return nil
}

// Processes begin-block events
func (p *Processor) processBeginBlockEvents(ctx context.Context, height int64, blockTime time.Time, events []abci.Event) error {
	return p.processEvents(ctx, height, blockTime, "", -1, events, "begin_block")
}

// Pprocesses end-block events
func (p *Processor) processEndBlockEvents(ctx context.Context, height int64, blockTime time.Time, events []abci.Event) error {
	return p.processEvents(ctx, height, blockTime, "", -1, events, "end_block")
}

// Processes txn events
func (p *Processor) processTxEvents(ctx context.Context, height int64, blockTime time.Time, txHash string, txIndex int, events []abci.Event) error {
	return p.processEvents(ctx, height, blockTime, txHash, txIndex, events, "tx")
}

// Processes events
func (p *Processor) processEvents(ctx context.Context, height int64, blockTime time.Time, txHash string, txIndex int, events []abci.Event, source string) error {
	for i, event := range events {
		module := extractModuleFromEventType(event.Type)

		var eventID string
		if txHash != "" {
			eventID = fmt.Sprintf("%s_%d_%s", txHash, i, event.Type)
		} else {
			eventID = fmt.Sprintf("%d_%s_%d_%s", height, source, i, event.Type)
		}

		attributes := make(models.Attributes, 0, len(event.Attributes))
		for _, attr := range event.Attributes {
			attributes = append(attributes, models.Attribute{
				Key:   string(attr.Key),
				Value: string(attr.Value),
			})
		}

		// Create event model
		eventModel := &models.Event{
			ID:          eventID,
			BlockHeight: height,
			BlockTime:   blockTime,
			TxHash:      txHash,
			TxIndex:     int32(txIndex),
			EventType:   event.Type,
			EventIndex:  int32(i),
			Module:      module,
			Attributes:  attributes,
			CreatedAt:   time.Now().UTC(),
		}

		// Save event to repo
		err := p.repository.SaveEvent(ctx, eventModel)
		if err != nil {
			return fmt.Errorf("failed to save event %s: %w", eventID, err)
		}
	}

	return nil
}

// Helpers
func getTxStatus(code uint32) string {
	if code == 0 {
		return "success"
	}
	return "failed"
}

func extractModuleFromEventType(eventType string) string {
	// TODO: Module extraction based on the event type
	return "unknown"
}

func extractSigners(tx types.Tx) ([]string, error) {
	// TODO: Extract Signer from transactions
	return []string{}, nil
}

func extractMessages(tx types.Tx) (models.Messages, error) {
	// TODO: Extract Messages from transactions
	return models.Messages{}, nil
}

func extractFee(tx types.Tx) string {
	// TODO: Extract Fee from transactions
	return "0"
}

func extractMemo(tx types.Tx) string {
	// TODO: Extract Memo from transactions
	return ""
}
