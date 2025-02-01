// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

type Transaction struct {
	ID        int                      `json:"ID"`
	TxHash    string                   `json:"TxHash"`
	BlockHash string                   `json:"BlockHash"`
	Sponsor   string                   `json:"Sponsor"`
	Actors    []string                 `json:"Actors"`
	Receivers []string                 `json:"Receivers"`
	MaxFee    float64                  `json:"MaxFee"`
	Success   bool                     `json:"Success"`
	Fee       uint64                   `json:"Fee"`
	Actions   []map[string]interface{} `json:"Actions"`
	Timestamp string                   `json:"Timestamp"`
}

type TransactionVolumes struct {
	Hours12 float64 `json:"12_hours"`
	Hours24 float64 `json:"24_hours"`
	Days7   float64 `json:"7_days"`
	Days30  float64 `json:"30_days"`
}

// CountFilteredTransactions counts transactions based on optional filters
func CountFilteredTransactions(db *sql.DB, txHash, blockHash, actionType, actionName, user string) (int, error) {
	query, args := buildTransactionFilterQuery("COUNT(*)", txHash, blockHash, actionType, actionName, user)
	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// FetchFilteredTransactions retrieves transactions based on optional filters
func FetchFilteredTransactions(db *sql.DB, txHash, blockHash, actionType, actionName, user, limit, offset string) ([]Transaction, error) {
	query, args := buildTransactionFilterQuery("*", txHash, blockHash, actionType, actionName, user)
	query = fmt.Sprintf("%s ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", query, len(args)+1, len(args)+2)

	args = append(args, limit, offset)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanTransactions(rows)
}

// Helper function to construct filter queries for transactions
func buildTransactionFilterQuery(selectFields, txHash, blockHash, actionType, actionName, user string) (string, []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM transactions WHERE 1=1", selectFields)
	args := []interface{}{}
	argCounter := 1

	// Filter by transaction hash
	if txHash != "" {
		query += fmt.Sprintf(" AND tx_hash ILIKE $%d", argCounter)
		args = append(args, "%"+txHash+"%")
		argCounter++
	}

	// Filter by block hash
	if blockHash != "" {
		query += fmt.Sprintf(" AND block_hash ILIKE $%d", argCounter)
		args = append(args, "%"+blockHash+"%")
		argCounter++
	}

	// Filter by action type
	if actionType != "" {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM actions
				WHERE actions.tx_hash = transactions.tx_hash
				AND actions.action_type = $%d
			)`, argCounter)
		args = append(args, actionType)
		argCounter++
	}

	// Filter by action name
	if actionName != "" {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM actions
				WHERE actions.tx_hash = transactions.tx_hash
				AND LOWER(actions.action_name) = $%d
			)`, argCounter)
		args = append(args, strings.ToLower(actionName))
		argCounter++
	}

	// Filter by user (sponsor or actor or receiver)
	if user != "" {
		query += fmt.Sprintf(`
			AND (
				sponsor ILIKE $%d
				OR EXISTS (
					SELECT 1
					FROM unnest(actors) AS actor
					WHERE actor ILIKE $%d
				)
				OR EXISTS (
					SELECT 1
					FROM unnest(receivers) AS receiver
					WHERE receiver ILIKE $%d
				)
			)`, argCounter, argCounter, argCounter)
		args = append(args, "%"+user+"%")
		argCounter++
	}

	return query, args
}

// FetchTransactionByHash retrieves a transaction by its hash
func FetchTransactionByHash(db *sql.DB, txHash string) (Transaction, error) {
	var tx Transaction
	var actionsJSON []byte

	err := db.QueryRow(`
        SELECT id, tx_hash, block_hash, sponsor, actors, receivers, max_fee, success, fee, actions, timestamp
        FROM transactions WHERE tx_hash = $1`, txHash).Scan(
		&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor,
		pq.Array(&tx.Actors), pq.Array(&tx.Receivers),
		&tx.MaxFee, &tx.Success, &tx.Fee, &actionsJSON, &tx.Timestamp)
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
			SELECT transactions.id, transactions.tx_hash, transactions.block_hash, transactions.sponsor, transactions.actors, transactions.receivers, transactions.max_fee, transactions.success, transactions.fee, transactions.actions, transactions.timestamp
			FROM transactions
			INNER JOIN blocks ON transactions.block_hash = blocks.block_hash
			WHERE blocks.block_height = $1`
	} else {
		// blockIdentifier is a block hash
		query = `
			SELECT id, tx_hash, block_hash, sponsor, actors, receivers, max_fee, success, fee, actions, timestamp
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

// FetchTransactionsByUser retrieves transactions by user (sponsor or actor or receiver) with pagination
func FetchTransactionsByUser(db *sql.DB, user, limit, offset string) ([]Transaction, error) {
	normalizedUser := "%" + strings.TrimPrefix(user, "0x") + "%"

	rows, err := db.Query(`
		SELECT id, tx_hash, block_hash, sponsor, actors, receivers, max_fee, success, fee, actions, timestamp
		FROM transactions
		WHERE sponsor ILIKE $1
		   OR EXISTS (
			   SELECT 1
			   FROM unnest(actors) AS actor
			   WHERE actor ILIKE $1
		   )
		   OR EXISTS (
			   SELECT 1
			   FROM unnest(receivers) AS receiver
			   WHERE receiver ILIKE $1
		   )
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`, normalizedUser, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanTransactions(rows)
}
// FetchTransactionVolumes retrieves transactions volumes by 12hrs, 24hrs, 7days & 30 days
func FetchTransactionVolumes(db *sql.DB) (TransactionVolumes, error) {
	var volumes TransactionVolumes

	query := `
        WITH transfer_actions AS (
            SELECT 
                a.timestamp,
                CAST(COALESCE((a.input->>'value'),'0') AS NUMERIC) as transfer_value
            FROM actions a
            WHERE a.action_type = 0  -- Transfer action type
            AND a.timestamp >= NOW() - $1::interval
        )
        SELECT COALESCE(SUM(transfer_value), 0) as total_volume
        FROM transfer_actions`

	// Get volumes for all our time periods
	err := db.QueryRow(query, "12 hours").Scan(&volumes.Hours12)
	if err != nil {
		return volumes, err
	}

	err = db.QueryRow(query, "24 hours").Scan(&volumes.Hours24)
	if err != nil {
		return volumes, err
	}

	err = db.QueryRow(query, "7 days").Scan(&volumes.Days7)
	if err != nil {
		return volumes, err
	}

	err = db.QueryRow(query, "30 days").Scan(&volumes.Days30)
	if err != nil {
		return volumes, err
	}

	return volumes, nil
}

// Helper function to scan transaction rows and unmarshal outputs
func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	var transactions []Transaction

	for rows.Next() {
		var tx Transaction
		var actionsJSON []byte
		if err := rows.Scan(
			&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor,
			pq.Array(&tx.Actors), pq.Array(&tx.Receivers),
			&tx.MaxFee, &tx.Success, &tx.Fee, &actionsJSON, &tx.Timestamp); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(actionsJSON, &tx.Actions); err != nil {
			return nil, errors.New("unable to parse actions")
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err() // Check for errors during iteration
}
