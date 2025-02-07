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
	ID          int                      `json:"ID"`
	TxHash      string                   `json:"TxHash"`
	BlockHash   string                   `json:"BlockHash"`
	BlockHeight int64                    `json:"BlockHeight"`
	Sponsor     string                   `json:"Sponsor"`
	Actors      []string                 `json:"Actors"`
	Receivers   []string                 `json:"Receivers"`
	MaxFee      float64                  `json:"MaxFee"`
	Success     bool                     `json:"Success"`
	Fee         uint64                   `json:"Fee"`
	Actions     []map[string]interface{} `json:"Actions"`
	Timestamp   string                   `json:"Timestamp"`
}

type TransactionVolumes struct {
	Hours12 float64 `json:"12_hours"`
	Hours24 float64 `json:"24_hours"`
	Days7   float64 `json:"7_days"`
	Days30  float64 `json:"30_days"`
}

type ActionVolumes struct {
	ActionType int     `json:"action_type"`
	ActionName string  `json:"action_name"`
	DataType   string  `json:"data_type"`
	Hours12    float64 `json:"12_hours"`
	Hours24    float64 `json:"24_hours"`
	Days7      float64 `json:"7_days"`
	Days30     float64 `json:"30_days"`
}

type TotalVolume struct {
	Total uint64 `json:"total"`
}

type ActionVolume struct {
	ActionType uint8  `json:"action_type"`
	ActionName string `json:"action_name"`
	TotalCount int64  `json:"total_count"`
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
	query = fmt.Sprintf("%s ORDER BY transactions.timestamp DESC LIMIT $%d OFFSET $%d",
		query, len(args)+1, len(args)+2)

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
	if selectFields == "COUNT(*)" {
		query := fmt.Sprintf("SELECT COUNT(*) FROM transactions LEFT JOIN blocks ON transactions.block_hash = blocks.block_hash WHERE 1=1")
		args := []interface{}{}
		return query, args
	}

	query := fmt.Sprintf("SELECT id, tx_hash, transactions.block_hash, blocks.block_height, sponsor, actors, receivers, max_fee, success, fee, actions, transactions.timestamp FROM transactions LEFT JOIN blocks ON transactions.block_hash = blocks.block_hash WHERE 1=1")
	args := []interface{}{}
	argCounter := 1

	// Filter by transaction hash
	if txHash != "" {
		query += fmt.Sprintf(" AND transactions.tx_hash ILIKE $%d", argCounter)
		args = append(args, "%"+txHash+"%")
		argCounter++
	}

	// Filter by block hash
	if blockHash != "" {
		query += fmt.Sprintf(" AND transactions.block_hash ILIKE $%d", argCounter)
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
        SELECT id, tx_hash, transactions.block_hash, blocks.block_height, sponsor, actors, receivers, max_fee, success, fee, actions, timestamp
		FROM transactions 
        LEFT JOIN blocks ON transactions.block_hash = blocks.block_hash 
        WHERE tx_hash = $1`, txHash).Scan(
		&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.BlockHeight, &tx.Sponsor,
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
			SELECT transactions.id, transactions.tx_hash, transactions.block_hash, blocks.block_height, transactions.sponsor, transactions.actors, transactions.receivers, transactions.max_fee, transactions.success, transactions.fee, transactions.actions, transactions.timestamp
			FROM transactions
			INNER JOIN blocks ON transactions.block_hash = blocks.block_hash
			WHERE blocks.block_height = $1`
	} else {
		// blockIdentifier is a block hash
		query = `
			SELECT id, tx_hash, transactions.block_hash, blocks.block_height, sponsor, actors, receivers, max_fee, success, fee, actions, timestamp
			FROM transactions
			LEFT JOIN blocks ON transactions.block_hash = blocks.block_hash
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
		SELECT id, tx_hash, transactions.block_hash, blocks.block_height, sponsor, actors, receivers, max_fee, success, fee, actions, timestamp
		FROM transactions
		LEFT JOIN blocks ON transactions.block_hash = blocks.block_hash
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

// FetchAllActionVolumes retrieves all actions volumes by 12hrs, 24hrs, 7days & 30 days
func FetchAllActionVolumes(db *sql.DB) ([]ActionVolumes, error) {
	var volumes []ActionVolumes

	// Always populate as new actions are implimented
	allActions := []struct {
		actionType int
		actionName string
		dataType   string
	}{
		{0, "Transfer", "value"},
		{4, "CreateAsset", "count"},
	}

	for _, action := range allActions {
		volume := ActionVolumes{
			ActionType: action.actionType,
			ActionName: action.actionName,
			DataType:   action.dataType,
		}

		var query string
		if action.dataType == "value" {
			query = `
                WITH action_volumes AS (
                    SELECT 
                        a.action_type,
                        a.action_name,
                        CAST(COALESCE((a.input->>'value'),'0') AS NUMERIC) as amount,
                        a.timestamp
                    FROM actions a
                    WHERE LOWER(a.action_name) = LOWER($1)
                    AND a.timestamp >= NOW() - $2::interval
                )
                SELECT COALESCE(SUM(amount), 0) as total
                FROM action_volumes`
		} else {
			query = `
                SELECT COUNT(*)
                FROM actions a
                WHERE LOWER(a.action_name) = LOWER($1)
                AND a.timestamp >= NOW() - $2::interval`
		}

		// Retreive volue for each periods we need
		intervals := []string{"12 hours", "24 hours", "7 days", "30 days"}
		for _, interval := range intervals {
			var total float64
			err := db.QueryRow(query, action.actionName, interval).Scan(&total)
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}

			switch interval {
			case "12 hours":
				volume.Hours12 = total
			case "24 hours":
				volume.Hours24 = total
			case "7 days":
				volume.Days7 = total
			case "30 days":
				volume.Days30 = total
			}
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// FetchTotalTransferVolume retrieves the all-time total transfer volume
func FetchTotalTransferVolume(db *sql.DB) (TotalVolume, error) {
	var volume TotalVolume

	query := `
        SELECT COALESCE(SUM(CAST(COALESCE((input->>'value'),'0') AS NUMERIC)), 0) as total
        FROM actions 
        WHERE action_type = 0 
        AND LOWER(action_name) = 'transfer'
    `

	err := db.QueryRow(query).Scan(&volume.Total)
	if err != nil && err != sql.ErrNoRows {
		return volume, err
	}

	return volume, nil
}

// FetchActionVolumesByName retrieves an actions volume by it's name
func FetchActionVolumesByName(db *sql.DB, actionName string) (ActionVolumes, error) {
	var volume ActionVolumes

	dataType := "value"
	if strings.ToLower(actionName) == "createasset" {
		dataType = "count"
	}

	var query string
	if dataType == "value" {
		query = `
            WITH action_volumes AS (
                SELECT 
                    a.action_type,
                    a.action_name,
                    CAST(COALESCE((a.input->>'value'),'0') AS NUMERIC) as amount,
                    a.timestamp
                FROM actions a
                WHERE LOWER(a.action_name) = LOWER($1)
                AND a.timestamp >= NOW() - $2::interval
            )
            SELECT 
                action_type,
                action_name,
                COALESCE(SUM(amount), 0) as total
            FROM action_volumes
            GROUP BY action_type, action_name`
	} else {
		query = `
            SELECT 
                a.action_type,
                a.action_name,
                COUNT(*) as total
            FROM actions a
            WHERE LOWER(a.action_name) = LOWER($1)
            AND a.timestamp >= NOW() - $2::interval
            GROUP BY action_type, action_name`
	}

	volume.DataType = dataType

	// Retreive volue for each periods we need
	intervals := []string{"12 hours", "24 hours", "7 days", "30 days"}
	for _, interval := range intervals {
		var total float64
		err := db.QueryRow(query, actionName, interval).Scan(
			&volume.ActionType,
			&volume.ActionName,
			&total,
		)
		if err != nil && err != sql.ErrNoRows {
			return volume, err
		}

		switch interval {
		case "12 hours":
			volume.Hours12 = total
		case "24 hours":
			volume.Hours24 = total
		case "7 days":
			volume.Days7 = total
		case "30 days":
			volume.Days30 = total
		}
	}

	return volume, nil
}

// FetchActionVolumes retrieves all action volumes
func FetchActionVolumes(db *sql.DB) ([]ActionVolume, error) {
	rows, err := db.Query(`
        SELECT action_type, action_name, total_count
        FROM action_volumes
        ORDER BY total_count DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totals []ActionVolume
	for rows.Next() {
		var total ActionVolume
		if err := rows.Scan(
			&total.ActionType,
			&total.ActionName,
			&total.TotalCount,
		); err != nil {
			return nil, err
		}
		totals = append(totals, total)
	}
	return totals, rows.Err()
}

// Helper function to scan transaction rows and unmarshal outputs
func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	var transactions []Transaction

	for rows.Next() {
		var tx Transaction
		var actionsJSON []byte
		if err := rows.Scan(
			&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.BlockHeight, &tx.Sponsor,
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
