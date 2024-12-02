package models

import (
	"database/sql"
	"strconv"
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
		return nil, err
	}
	defer rows.Close()

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

// FetchBlock retrieves a block by its height or hash.
func FetchBlock(db *sql.DB, identifier string) (Block, error) {
	var block Block
	var whereClause string

	// Determine if identifier is a block height (integer) or block hash (string)
	if _, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		whereClause = "block_height = $1::bigint"
	} else {
		whereClause = "block_hash = $1"
	}

	// Base query with dynamic WHERE clause
	query := `
        SELECT block_height, block_hash, parent_block_hash, state_root, block_size, tx_count, total_fee, avg_tx_size, unique_participants, timestamp
        FROM blocks
        WHERE ` + whereClause

	// Execute query with identifier as parameter
	err := db.QueryRow(query, identifier).Scan(&block.BlockHeight, &block.BlockHash, &block.ParentBlockHash, &block.StateRoot, &block.BlockSize, &block.TxCount, &block.TotalFee, &block.AvgTxSize, &block.UniqueParticipants, &block.Timestamp)

	return block, err
}
