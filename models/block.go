// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql"
	"fmt"
	"log"
)

type Block struct {
	BlockHeight        int64   `json:"BlockHeight"`
	BlockHash          string  `json:"BlockHash"`
	ParentBlockHash    string  `json:"ParentBlockHash"`
	StateRoot          string  `json:"StateRoot"`
	BlockSize          int     `json:"BlockSize"`
	TxCount            int     `json:"TxCount"`
	TotalFee           float64 `json:"TotalFee"`
	AvgTxSize          float64 `json:"AvgTxSize"`
	UniqueParticipants int     `json:"UniqueParticipants"`
	Timestamp          string  `json:"Timestamp"`
}

// FetchAllBlocks retrieves blocks from the database with pagination
func FetchAllBlocks(db *sql.DB, limit, offset string) ([]Block, error) {
	rows, err := db.Query(`SELECT * FROM blocks ORDER BY block_height DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanBlocks(rows)
}

// FetchBlock retrieves a block by its height or hash.
func FetchBlock(db *sql.DB, height, hash string) (Block, error) {
	var block Block
	var query string
	var param interface{}

	// Determine if identifier is a block height (integer) or block hash (string)
	if height != "" {
		query = `SELECT * FROM blocks WHERE block_height = $1::bigint`
		param = height
	} else if hash != "" {
		query = `SELECT * FROM blocks WHERE block_hash = $1`
		param = hash
	} else {
		return block, fmt.Errorf("either block height or hash must be provided")
	}

	err := db.QueryRow(query, param).Scan(
		&block.BlockHeight, &block.BlockHash, &block.ParentBlockHash,
		&block.StateRoot, &block.BlockSize, &block.TxCount, &block.TotalFee,
		&block.AvgTxSize, &block.UniqueParticipants, &block.Timestamp)

	return block, err
}

// Helper function to scan block rows
func scanBlocks(rows *sql.Rows) ([]Block, error) {
	var blocks []Block

	for rows.Next() {
		var block Block
		if err := rows.Scan(&block.BlockHeight, &block.BlockHash, &block.ParentBlockHash, &block.StateRoot, &block.BlockSize, &block.TxCount, &block.TotalFee, &block.AvgTxSize, &block.UniqueParticipants, &block.Timestamp); err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}
