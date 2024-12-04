package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
)

type Transaction struct {
	ID        int                      `json:"ID"`
	TxHash    string                   `json:"TxHash"`
	BlockHash string                   `json:"BlockHash"`
	Sponsor   string                   `json:"Sponsor"`
	MaxFee    float64                  `json:"MaxFee"`
	Success   bool                     `json:"Success"`
	Fee       uint64                   `json:"Fee"`
	Actions   []map[string]interface{} `json:"Actions"`
	Timestamp string                   `json:"Timestamp"`
}

// FetchAllTransactions retrieves transactions from the database with pagination
func FetchAllTransactions(db *sql.DB, limit, offset string) ([]Transaction, error) {
	rows, err := db.Query(`SELECT * FROM transactions ORDER BY timestamp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// FetchTransactionByHash retrieves a transaction by its hash
func FetchTransactionByHash(db *sql.DB, txHash string) (Transaction, error) {
	var tx Transaction
	var actionsJSON []byte

	err := db.QueryRow(`SELECT * FROM transactions WHERE tx_hash = $1`, txHash).Scan(
		&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.MaxFee, &tx.Success, &tx.Fee, &actionsJSON, &tx.Timestamp)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return tx, err
	}

	if err := json.Unmarshal(actionsJSON, &tx.Actions); err != nil {
		return tx, err
	}

	return tx, nil
}

func FetchTransactionsByBlock(db *sql.DB, blockIdentifier string) ([]Transaction, error) {
	// Determine the query based on whether blockIdentifier is numeric
	var query string

	if _, err := strconv.ParseInt(blockIdentifier, 10, 64); err == nil {
		// blockIdentifier is a block height
		query = `
			SELECT transactions.id, transactions.tx_hash, transactions.block_hash, transactions.sponsor, transactions.max_fee, transactions.success, transactions.fee, transactions.actions, transactions.timestamp
			FROM transactions
			INNER JOIN blocks ON transactions.block_hash = blocks.block_hash
			WHERE blocks.block_height = $1`
	} else {
		// blockIdentifier is a block hash
		query = `
			SELECT id, tx_hash, block_hash, sponsor, max_fee, success, fee, actions, timestamp
			FROM transactions
			WHERE block_hash = $1`
	}

	// Execute the query
	rows, err := db.Query(query, blockIdentifier)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// FetchTransactionsByUser retrieves transactions by user (sponsor) with pagination
func FetchTransactionsByUser(db *sql.DB, user, limit, offset string) ([]Transaction, error) {
	rows, err := db.Query(`
        SELECT * FROM transactions
        WHERE sponsor = $1
        ORDER BY timestamp DESC
        LIMIT $2 OFFSET $3
    `, user, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// Helper function to scan transaction rows and unmarshal outputs
func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	var transactions []Transaction

	for rows.Next() {
		var tx Transaction
		var actionsJSON []byte
		if err := rows.Scan(&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.MaxFee, &tx.Success, &tx.Fee, &actionsJSON, &tx.Timestamp); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(actionsJSON, &tx.Actions); err != nil {
			return nil, errors.New("unable to parse actions")
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err() // Check for errors during iteration
}
