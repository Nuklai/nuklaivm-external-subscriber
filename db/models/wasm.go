// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"encoding/json"
	"time"
)

type WasmCode struct {
	CodeID            int64  `db:"code_id" json:"code_id"`
	Creator           string `db:"creator" json:"creator"`
	DataHash          string `db:"data_hash" json:"data_hash"`
	UploadTxHash      string `db:"upload_tx_hash" json:"upload_tx_hash"`
	InstantiatedCount int64  `db:"instantiated_count" json:"instantiated_count"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// An Init WASM contract
type WasmContract struct {
	Address           string          `db:"address" json:"address"`
	CodeID            int64           `db:"code_id" json:"code_id"`
	Creator           string          `db:"creator" json:"creator"`
	Admin             string          `db:"admin" json:"admin"`
	Label             string          `db:"label" json:"label"`
	CreatedHeight     int64           `db:"created_height" json:"created_height"`
	InstantiateTxHash string          `db:"instantiate_tx_hash" json:"instantiate_tx_hash"`
	InitMsg           json.RawMessage `db:"init_msg" json:"init_msg"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// WASM contract execution
type WasmExecute struct {
	ID              string          `db:"id" json:"id"`
	ContractAddress string          `db:"contract_address" json:"contract_address"`
	Sender          string          `db:"sender" json:"sender"`
	TxHash          string          `db:"tx_hash" json:"tx_hash"`
	BlockHeight     int64           `db:"block_height" json:"block_height"`
	BlockTime       time.Time       `db:"block_time" json:"block_time"`
	ExecuteMsg      json.RawMessage `db:"execute_msg" json:"execute_msg"`
	Funds           Coins           `db:"funds" json:"funds"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// WASM contract query
type WasmQuery struct {
	ID              string          `db:"id" json:"id"`
	ContractAddress string          `db:"contract_address" json:"contract_address"`
	Sender          string          `db:"sender" json:"sender"`
	QueryMsg        json.RawMessage `db:"query_msg" json:"query_msg"`
	Response        json.RawMessage `db:"response" json:"response"`
	BlockHeight     int64           `db:"block_height" json:"block_height"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
