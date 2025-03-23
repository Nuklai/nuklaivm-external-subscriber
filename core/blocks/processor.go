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
	"github.com/nuklai/nuklaivm-external-subscriber/core/parser"
	"github.com/nuklai/nuklaivm-external-subscriber/core/state"
	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
	"github.com/nuklai/nuklaivm-external-subscriber/db/repository"
	"github.com/nuklai/nuklaivm-external-subscriber/metrics"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Processor is in charge of processing blocks
type Processor struct {
	fetcher        *Fetcher
	repository     repository.Repository
	txParser       *parser.TxParser
	stateExtractor state.Extractor
	rpcClient      *rpc.Client
}

func NewProcessor(fetcher *Fetcher, repo repository.Repository, rpcClient *rpc.Client) *Processor {
	return &Processor{
		fetcher:        fetcher,
		repository:     repo,
		txParser:       parser.NewTxParser(),
		rpcClient:      rpcClient,
		stateExtractor: state.NewStateExtractor(rpcClient, repo),
	}
}

// Process a block at the specified height
func (p *Processor) ProcessBlock(ctx context.Context, height int64) error {
	startTime := time.Now()

	// Get the block from the RPC
	blockResult, blockResultsResult, err := p.fetcher.FetchBlock(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to fetch block at height %d: %w", height, err)
	}

	// Get the block results from the RPC
	_, err = p.processAndSaveBlock(ctx, blockResult, blockResultsResult)
	if err != nil {
		return fmt.Errorf("failed to process and save block at height %d: %w", height, err)
	}

	// Process and save transactions
	err = p.processAndSaveTransactions(ctx, blockResult, blockResultsResult)
	if err != nil {
		return fmt.Errorf("failed to process and save transactions at height %d: %w", height, err)
	}

	// Process finalize-block events
	err = p.processFinalizeBlockEvents(ctx, height, blockResult.Block.Time, blockResultsResult.FinalizeBlockEvents)
	if err != nil {
		return fmt.Errorf("failed to process finalize-block events at height %d: %w", height, err)
	}

	// Process the state extraction for this block
	err = p.processStateExtraction(ctx, height, blockResult.Block.Time)
	if err != nil {
		logger.Warn("Failed to extract state at height", zap.Int64("height", height), zap.Error(err))
	}

	// Update metrics
	metrics.BlocksProcessed.Inc()
	metrics.BlockProcessingTime.Observe(time.Since(startTime).Seconds())
	metrics.TransactionsProcessed.Add(float64(len(blockResult.Block.Txs)))

	logger.Info("Processed block",
		zap.Int64("height", height),
		zap.Time("time", blockResult.Block.Time),
		zap.Int("num_txs", len(blockResult.Block.Txs)),
		zap.Duration("duration", time.Since(startTime)),
	)

	return nil
}

// Save block to the repo
func (p *Processor) processAndSaveBlock(ctx context.Context, blockResult *ctypes.ResultBlock, blockResultsResult *ctypes.ResultBlockResults) (*models.Block, error) {
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
	err := p.repository.SaveBlock(ctx, blockModel)
	if err != nil {
		return nil, fmt.Errorf("failed to save block: %w", err)
	}

	return blockModel, nil
}

// Process and savea the txn to the repo
func (p *Processor) processAndSaveTransactions(ctx context.Context, blockResult *ctypes.ResultBlock, blockResultsResult *ctypes.ResultBlockResults) error {
	block := blockResult.Block
	txResults := blockResultsResult.TxResults

	// Process each transaction
	for i, txBytes := range block.Txs {
		txHash := hex.EncodeToString(txBytes.Hash())
		txResult := txResults[i]

		// Get transaction details
		txDetail, err := p.fetcher.FetchTransaction(ctx, txBytes.Hash())
		if err != nil {
			logger.Warn("Failed to fetch transaction details",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			continue
		}

		// Extract signers from transaction
		signers, err := p.extractSigners(txDetail.Tx)
		if err != nil {
			logger.Warn("Failed to extract signers from transaction",
				zap.String("hash", txHash),
				zap.Error(err),
			)
			signers = []string{}
		}

		// Extract messages from transaction
		messages, err := p.extractMessages(txDetail.Tx)
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
			Fee:         p.extractFee(txDetail.Tx),
			Memo:        p.extractMemo(txDetail.Tx),
			ErrorLog:    txResult.Log,
			Status:      getTxStatus(txResult.Code),
			Codespace:   txResult.Codespace,
			Messages:    messages,
			Signers:     models.Signers(signers),
			CreatedAt:   time.Now().UTC(),
		}

		// Save txn to repo
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

// Process finalize-block events
func (p *Processor) processFinalizeBlockEvents(ctx context.Context, height int64, blockTime time.Time, events []abci.Event) error {
	return p.processEvents(ctx, height, blockTime, "", -1, events, "finalize_block")
}

// Process txn events
func (p *Processor) processTxEvents(ctx context.Context, height int64, blockTime time.Time, txHash string, txIndex int, events []abci.Event) error {
	return p.processEvents(ctx, height, blockTime, txHash, txIndex, events, "tx")
}

// Process events
func (p *Processor) processEvents(ctx context.Context, height int64, blockTime time.Time, txHash string, txIndex int, events []abci.Event, source string) error {
	for i, event := range events {
		module := p.extractModuleFromEventType(event.Type)

		var eventID string
		if txHash != "" {
			eventID = fmt.Sprintf("%s_%d_%s", txHash, i, event.Type)
		} else {
			eventID = fmt.Sprintf("%d_%s_%d_%s", height, source, i, event.Type)
		}

		attributes := make(models.Attributes, 0, len(event.Attributes))
		for _, attr := range event.Attributes {
			attributes = append(attributes, models.Attribute{
				Key:   attr.Key,
				Value: attr.Value,
			})
		}

		// Create event model
		eventModel := &models.Event{
			ID:          eventID,
			BlockHeight: height,
			BlockTime:   blockTime,
			TxIndex:     int32(txIndex),
			EventType:   event.Type,
			EventIndex:  int32(i),
			Module:      module,
			Attributes:  attributes,
			CreatedAt:   time.Now().UTC(),
		}

		if txHash != "" {
			eventModel.TxHash = txHash
		}

		// Save event to repo
		err := p.repository.SaveEvent(ctx, eventModel)
		if err != nil {
			return fmt.Errorf("failed to save event %s: %w", eventID, err)
		}

		// Process module-specific events
		err = p.processModuleSpecificEvent(ctx, height, blockTime, eventModel)
		if err != nil {
			logger.Warn("Failed to process module-specific event",
				zap.String("id", eventID),
				zap.String("module", module),
				zap.Error(err),
			)
			// ... other events
		}
	}

	return nil
}

func (p *Processor) processModuleSpecificEvent(ctx context.Context, height int64, blockTime time.Time, event *models.Event) error {
	switch event.Module {
	case "dataset":
		return p.processDatasetEvent(ctx, event)
	case "staking":
		return p.processStakingEvent(ctx, event)
	case "gov":
		return p.processGovEvent(ctx, event)
	case "wasm":
		return p.processWasmEvent(ctx, event)
	case "bank":
		return p.processBankEvent(ctx, event)
	default:
		return nil
	}
}

// Process dataset-specific events
func (p *Processor) processDatasetEvent(ctx context.Context, event *models.Event) error {
	return nil
}

// Process staking-specific events
func (p *Processor) processStakingEvent(ctx context.Context, event *models.Event) error {
	return nil
}

// Process governance-specific events
func (p *Processor) processGovEvent(ctx context.Context, event *models.Event) error {
	return nil
}

// Process WASM-specific events
func (p *Processor) processWasmEvent(ctx context.Context, event *models.Event) error {
	return nil
}

// Process bank-specific events
func (p *Processor) processBankEvent(ctx context.Context, event *models.Event) error {
	return nil
}

// Extract and save state data for the current block
func (p *Processor) processStateExtraction(ctx context.Context, height int64, blockTime time.Time) error {
	req := &state.StateRequest{
		Height:    height,
		BlockTime: blockTime,
		// AccountAddresses: []string{},
	}

	stateResponse, err := p.stateExtractor.ExtractState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to extract state: %w", err)
	}

	for _, validator := range stateResponse.Validators {
		err := p.repository.SaveValidator(ctx, validator)
		if err != nil {
			logger.Warn("Failed to save validator",
				zap.String("address", validator.OperatorAddress),
				zap.Error(err),
			)
		}
	}

	for _, account := range stateResponse.Accounts {
		err := p.repository.SaveAccount(ctx, account)
		if err != nil {
			logger.Warn("Failed to save account",
				zap.String("address", account.Address),
				zap.Error(err),
			)
		}
	}

	for _, delegation := range stateResponse.Delegations {
		err := p.repository.SaveDelegation(ctx, delegation)
		if err != nil {
			logger.Warn("Failed to save delegation",
				zap.String("id", delegation.ID),
				zap.Error(err),
			)
		}
	}

	for _, dataset := range stateResponse.Datasets {
		err := p.repository.SaveDataset(ctx, dataset)
		if err != nil {
			logger.Warn("Failed to save dataset",
				zap.String("denom", dataset.Denom),
				zap.Error(err),
			)
		}
	}

	for _, contract := range stateResponse.WasmContracts {
		err := p.repository.SaveWasmContract(ctx, contract)
		if err != nil {
			logger.Warn("Failed to save WASM contract",
				zap.String("address", contract.Address),
				zap.Error(err),
			)
		}
	}

	for _, proposal := range stateResponse.Proposals {
		err := p.repository.SaveProposal(ctx, proposal)
		if err != nil {
			logger.Warn("Failed to save proposal",
				zap.Int64("id", proposal.ProposalID),
				zap.Error(err),
			)
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

func (p *Processor) extractModuleFromEventType(eventType string) string {
	return parser.MapEventTypeToModule(eventType)
}

func (p *Processor) extractSigners(tx types.Tx) ([]string, error) {
	_, signers, _, err := p.txParser.ParseTx(tx)
	return signers, err
}

func (p *Processor) extractMessages(tx types.Tx) (models.Messages, error) {
	messages, _, _, err := p.txParser.ParseTx(tx)
	return messages, err
}

func (p *Processor) extractFee(tx types.Tx) string {
	fee, err := p.txParser.ExtractFee(tx)
	if err != nil {
		return "0"
	}
	return fee
}

func (p *Processor) extractMemo(tx types.Tx) string {
	_, _, memo, err := p.txParser.ParseTx(tx)
	if err != nil {
		return ""
	}
	return memo
}
