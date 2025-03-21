// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package blocks

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
	"github.com/nuklai/nuklaivm-external-subscriber/client/rpc"
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
func (f *Fetcher) FetchBlock(ctx context.Context, height int64) (*ctypes.ResultBlock, *ctypes.ResultBlockResults, error) {
	// Get block from RPC
	blockResult, err := f.rpcClient.Block(ctx, height)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	// Get the block results
	blockResultsResult, err := f.rpcClient.BlockResults(ctx, height)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block results at height %d: %w", height, err)
	}

	return blockResult, blockResultsResult, nil
}

// Fetch a block and its txns at the specified height
func (f *Fetcher) FetchTransaction(ctx context.Context, hash []byte) (*ctypes.ResultTx, error) {
	txDetail, err := f.rpcClient.Tx(ctx, hash, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction details for hash %X: %w", hash, err)
	}

	return txDetail, nil
}

// Fetch validators at the specified height
func (f *Fetcher) FetchValidators(ctx context.Context, height int64) ([]*types.Validator, error) {
	page := 1
	perPage := 100
	validators := make([]*types.Validator, 0)

	for {
		// Get validators from RPC
		validatorsResult, err := f.rpcClient.Validators(ctx, height, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to get validators at height %d (page %d): %w", height, page, err)
		}

		validators = append(validators, validatorsResult.Validators...)

		if len(validatorsResult.Validators) < perPage {
			break
		}

		page++
	}

	return validators, nil
}

// Performs an ABCI query
func (f *Fetcher) FetchABCIQuery(ctx context.Context, path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	result, err := f.rpcClient.ABCIQuery(ctx, path, data)
	if err != nil {
		logger.Debug("ABCI query failed",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to execute ABCI query %s: %w", path, err)
	}

	if result.Response.Code != 0 {
		logger.Debug("ABCI query returned non-zero code",
			zap.String("path", path),
			zap.Uint32("code", result.Response.Code),
			zap.String("log", result.Response.Log),
		)
		return result, fmt.Errorf("ABCI query failed with code %d: %s", result.Response.Code, result.Response.Log)
	}

	return result, nil
}

func (f *Fetcher) FetchEvents(ctx context.Context, height int64) ([]abci.Event, [][]abci.Event, error) {
	_, blockResults, err := f.FetchBlock(ctx, height)
	if err != nil {
		return nil, nil, err
	}

	blockEvents := blockResults.FinalizeBlockEvents
	txEvents := make([][]abci.Event, len(blockResults.TxResults))

	for i, txResult := range blockResults.TxResults {
		txEvents[i] = txResult.Events
	}

	return blockEvents, txEvents, nil
}
