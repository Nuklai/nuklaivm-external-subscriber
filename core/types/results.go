// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package core_types

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/types"
)

// Contains the block information
type ResultBlock struct {
	BlockID types.BlockID `json:"block_id"`
	Block   *types.Block  `json:"block"`
}

// Contains results from block execution
type ResultBlockResults struct {
	Height                int64                  `json:"height"`
	TxsResults            []*abci.ExecTxResult   `json:"txs_results"`
	BeginBlockEvents      []abci.Event           `json:"begin_block_events"`
	EndBlockEvents        []abci.Event           `json:"end_block_events"`
	ValidatorUpdates      []abci.ValidatorUpdate `json:"validator_updates"`
	ConsensusParamUpdates *abci.ConsensusParams  `json:"consensus_param_updates"`
}

// Contains a node status info
type ResultStatus struct {
	NodeInfo      types.NodeInfo `json:"node_info"`
	SyncInfo      SyncInfo       `json:"sync_info"`
	ValidatorInfo ValidatorInfo  `json:"validator_info"`
}

// Contains sync status info
type SyncInfo struct {
	LatestBlockHash   bytes.HexBytes `json:"latest_block_hash"`
	LatestAppHash     bytes.HexBytes `json:"latest_app_hash"`
	LatestBlockHeight int64          `json:"latest_block_height"`
	LatestBlockTime   time.Time      `json:"latest_block_time"`
	CatchingUp        bool           `json:"catching_up"`
}

// Contains validator info
type ValidatorInfo struct {
	Address     bytes.HexBytes `json:"address"`
	PubKey      []byte         `json:"pub_key"`
	VotingPower int64          `json:"voting_power"`
}

// Contains info about a transaction
type ResultTx struct {
	Hash     bytes.HexBytes    `json:"hash"`
	Height   int64             `json:"height"`
	Index    uint32            `json:"index"`
	TxResult abci.ExecTxResult `json:"tx_result"`
	Tx       types.Tx          `json:"tx"`
	Proof    types.TxProof     `json:"proof,omitempty"`
}

// Contains search results for a txn
type ResultTxSearch struct {
	Txs        []*ResultTx `json:"txs"`
	TotalCount int         `json:"total_count"`
}

// Contains info about the validator set
type ResultValidators struct {
	BlockHeight int64              `json:"block_height"`
	Validators  []*types.Validator `json:"validators"`
	Count       int                `json:"count"`
	Total       int                `json:"total"`
}

// Contains results of an ABCI query
type ResultABCIQuery struct {
	Response abci.ResponseQuery `json:"response"`
}

// Contains result after broadcasting a txn
type ResultBroadcastTx struct {
	Code      uint32         `json:"code"`
	Data      bytes.HexBytes `json:"data"`
	Log       string         `json:"log"`
	Codespace string         `json:"codespace"`
	Hash      bytes.HexBytes `json:"hash"`
}
