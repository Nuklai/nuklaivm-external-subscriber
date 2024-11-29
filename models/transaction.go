package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
)

type Transaction struct {
	ID        int                    `json:"ID"`
	TxHash    string                 `json:"TxHash"`
	BlockHash string                 `json:"BlockHash"`
	Sponsor   string                 `json:"Sponsor"`
	Sender    string                 `json:"Sender"`
    Receiver  string                 `json:"Receiver"`
	MaxFee    float64                `json:"MaxFee"`
	Success   bool                   `json:"Success"`
	Fee       uint64                 `json:"Fee"`
	Outputs   map[string]interface{} `json:"Outputs"`
	Timestamp string                 `json:"Timestamp"`
}

// FetchAllTransactions retrieves transactions from the database with pagination
func FetchAllTransactions(db *sql.DB, limit, offset string) ([]Transaction, error) {
	rows, err := db.Query(`SELECT * FROM transactions ORDER BY timestamp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		var outputsJSON []byte
		if err := rows.Scan(&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.Sender, &tx.Receiver, &tx.MaxFee, &tx.Success, &tx.Fee, &outputsJSON, &tx.Timestamp); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(outputsJSON, &tx.Outputs); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// FetchTransactionByHash retrieves a transaction by its hash
func FetchTransactionByHash(db *sql.DB, txHash string) (Transaction, error) {
	var tx Transaction
	var outputsJSON []byte

	err := db.QueryRow(`SELECT * FROM transactions WHERE tx_hash = $1`, txHash).Scan(
		&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.Sender, &tx.Receiver, &tx.MaxFee, &tx.Success, &tx.Fee, &outputsJSON, &tx.Timestamp)
	if err != nil {
		return tx, err
	}

	if err := json.Unmarshal(outputsJSON, &tx.Outputs); err != nil {
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
			SELECT transactions.id, transactions.tx_hash, transactions.block_hash, transactions.sponsor, transactions.sender, transactions.receiver, transactions.max_fee, transactions.success, transactions.fee, transactions.outputs, transactions.timestamp
			FROM transactions
			INNER JOIN blocks ON transactions.block_hash = blocks.block_hash
			WHERE blocks.block_height = $1`
	} else {
		// blockIdentifier is a block hash
		query = `
			SELECT id, tx_hash, block_hash, sponsor, max_fee, success, fee, outputs, timestamp
			FROM transactions
			WHERE block_hash = $1`
	}

	// Execute the query
	rows, err := db.Query(query, blockIdentifier)
	if err != nil {
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
		var outputsJSON []byte
		if err := rows.Scan(&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.Sender, &tx.Receiver, &tx.MaxFee, &tx.Success, &tx.Fee, &outputsJSON, &tx.Timestamp); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(outputsJSON, &tx.Outputs); err != nil {
			return nil, errors.New("unable to parse outputs")
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err() // Check for errors during iteration
}
