// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import "time"

type Block struct {
	Height          int64     `db:"height" json:"height"`
	Hash            string    `db:"hash" json:"hash"`
	ProposerAddress string    `db:"proposer_address" json:"proposer_address"`
	NumTxs          int64     `db:"num_txs" json:"num_txs"`
	TotalGas        int64     `db:"total_gas" json:"total_gas"`
	Size            int64     `db:"size" json:"size"`
	Time            time.Time `db:"time" json:"time"`
	AppHash         string    `db:"app_hash" json:"app_hash"`
	Version         int64     `db:"version" json:"version"`
	ChainID         string    `db:"chain_id" json:"chain_id"`

	// Prev block info
	LastBlockID    string `db:"last_block_id" json:"last_block_id"`
	LastCommitHash string `db:"last_commit_hash" json:"last_commit_hash"`

	// Consensus info
	ValidatorsHash     string `db:"validators_hash" json:"validators_hash"`
	NextValidatorsHash string `db:"next_validators_hash" json:"next_validators_hash"`
	ConsensusHash      string `db:"consensus_hash" json:"consensus_hash"`

	// Data hashes
	DataHash     string `db:"data_hash" json:"data_hash"`
	EvidenceHash string `db:"evidence_hash" json:"evidence_hash"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
