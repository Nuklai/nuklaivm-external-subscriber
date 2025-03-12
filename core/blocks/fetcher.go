// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/types"
	"github.com/nuklai/nuklaivm-external-subscriber/client/rpc"
	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Fetches data from the VM
type Fetcher struct {
	rpcClient *rpc.Client
}

func NewFetcher(rpcClient *rpc.Client) *Fetcher {
	return &Fetcher{
		rpcClient: rpcClient,
	}
}

// Fetch block at the specified height
func (f *Fetcher) FetchBlock(ctx context.Context, height int64) (*models.Block, error) {
	blockResult, err := f.rpcClient.Block(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	blockResultsResult, err := f.rpcClient.BlockResults(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block results at height %d: %w", height, err)
	}

	totalGas := int64(0)
	for _, txResult := range blockResultsResult.TxResults {
		totalGas += txResult.GasUsed
	}

	block := blockResult.Block

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

	return blockModel, nil
}

// Fetch a block and its transactions at the specified height
func (f *Fetcher) FetchBlockWithTransactions(ctx context.Context, height int64) (*models.Block, []*models.Transaction, error) {
	blockResult, err := f.rpcClient.Block(ctx, height)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	blockResultsResult, err := f.rpcClient.BlockResults(ctx, height)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block results at height %d: %w", height, err)
	}

	// Get block model frm FetchBlock
	blockModel, err := f.FetchBlock(ctx, height)
	if err != nil {
		return nil, nil, err
	}

	transactions := make([]*models.Transaction, 0, len(blockResult.Block.Txs))
	for i, txBytes := range blockResult.Block.Txs {
		txHash := hex.EncodeToString(txBytes.Hash())
		txResult := blockResultsResult.TxResults[i]

		// Get the transaction details
		txDetail, err := f.rpcClient.Tx(ctx, txBytes.Hash(), false)
		if err != nil {
			logger.Warn("Failed to get transaction details",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			continue
		}

		// Extract signers
		signers, err := f.extractSigners(txDetail.Tx)
		if err != nil {
			logger.Warn("Failed to extract signers from transaction",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			signers = []string{}
		}

		// Extract messages
		messages, err := f.extractMessages(txDetail.Tx)
		if err != nil {
			logger.Warn("Failed to extract messages from transaction",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			messages = models.Messages{}
		}

		// Create transaction model
		txModel := &models.Transaction{
			Hash:        txHash,
			BlockHeight: blockResult.Block.Height,
			BlockTime:   blockResult.Block.Time,
			Index:       int32(i),
			GasWanted:   txResult.GasWanted,
			GasUsed:     txResult.GasUsed,
			Fee:         f.extractFee(txDetail.Tx),
			Memo:        f.extractMemo(txDetail.Tx),
			ErrorLog:    txResult.Log,
			Status:      f.getTxStatus(txResult.Code),
			Codespace:   txResult.Codespace,
			Messages:    messages,
			Signers:     models.Signers(signers),
			CreatedAt:   time.Now().UTC(),
		}

		transactions = append(transactions, txModel)
	}

	return blockModel, transactions, nil
}

// Fetch events for a block at the specified height
func (f *Fetcher) FetchEvents(ctx context.Context, height int64) ([]*models.Event, error) {
	blockResult, err := f.rpcClient.Block(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	blockResultsResult, err := f.rpcClient.BlockResults(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block results at height %d: %w", height, err)
	}

	events := make([]*models.Event, 0)

	// Process finalize-block events
	finalizeBlockEvents, err := f.processEvents(height, blockResult.Block.Time, "", -1, blockResultsResult.FinalizeBlockEvents, "finalize_block")
	if err != nil {
		return nil, fmt.Errorf("failed to process finalize-block events: %w", err)
	}
	events = append(events, finalizeBlockEvents...)

	// Process tnx events
	for i, txBytes := range blockResult.Block.Txs {
		txHash := hex.EncodeToString(txBytes.Hash())
		txResult := blockResultsResult.TxResults[i]

		txEvents, err := f.processEvents(height, blockResult.Block.Time, txHash, i, txResult.Events, "tx")
		if err != nil {
			return nil, fmt.Errorf("failed to process tx events: %w", err)
		}
		events = append(events, txEvents...)
	}

	return events, nil
}

// Fetch validators at the specified height
func (f *Fetcher) FetchValidators(ctx context.Context, height int64) ([]*models.Validator, error) {
	page := 1
	perPage := 100
	validators := make([]*models.Validator, 0)

	for {
		validatorsResult, err := f.rpcClient.Validators(ctx, height, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to get validators at height %d (page %d): %w", height, page, err)
		}

		for _, val := range validatorsResult.Validators {
			consensusAddr := hex.EncodeToString(val.Address)

			operatorAddr := consensusAddr
			accountAddr := consensusAddr

			votingPower := val.VotingPower
			if votingPower < 0 {
				votingPower = 0
			}

			validator := &models.Validator{
				ConsensusAddress: consensusAddr,
				OperatorAddress:  operatorAddr,
				AccountAddress:   accountAddr,
				// Moniker:          val.Name,
				Identity:        "",
				Website:         "",
				SecurityContact: "",
				Details:         "",
				CommissionRate:  "0",
				VotingPower:     strconv.FormatInt(votingPower, 10),
				Status:          f.getValidatorStatus(val),
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			}

			validators = append(validators, validator)
		}

		if len(validatorsResult.Validators) < perPage {
			break
		}

		page++
	}

	return validators, nil
}

// Helpers
func (f *Fetcher) getTxStatus(code uint32) string {
	if code == 0 {
		return "success"
	}
	return "failed"
}

func (f *Fetcher) getValidatorStatus(validator *types.Validator) string {
	if validator.VotingPower <= 0 {
		return "inactive"
	}
	return "active"
}

func (f *Fetcher) processEvents(height int64, blockTime time.Time, txHash string, txIndex int, events []abci.Event, source string) ([]*models.Event, error) {
	result := make([]*models.Event, 0, len(events))

	for i, event := range events {
		module := f.extractModuleFromEventType(event.Type)

		// Event ID
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

		result = append(result, eventModel)
	}

	return result, nil
}

func (f *Fetcher) extractModuleFromEventType(eventType string) string {
	for i, c := range eventType {
		if c == '.' {
			return eventType[:i]
		}
	}
	return "unknown"
}

// TODOs
func (f *Fetcher) extractSigners(tx types.Tx) ([]string, error) {
	return []string{}, nil
}

func (f *Fetcher) extractMessages(tx types.Tx) (models.Messages, error) {
	return models.Messages{}, nil
}

func (f *Fetcher) extractFee(tx types.Tx) string {
	return "0"
}

func (f *Fetcher) extractMemo(tx types.Tx) string {
	return ""
}
