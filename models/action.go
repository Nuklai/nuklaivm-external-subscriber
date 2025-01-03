// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
)

type Action struct {
	ID          int                    `json:"ID"`
	TxHash      string                 `json:"TxHash"`
	ActionType  int                    `json:"ActionType"`
	ActionName  string                 `json:"ActionName"`
	ActionIndex int                    `json:"ActionIndex"`
	Input       map[string]interface{} `json:"Input"`
	Output      map[string]interface{} `json:"Output"`
	Timestamp   string                 `json:"Timestamp"`
}

// FetchAllActions retrieves actions from the database with pagination
func FetchAllActions(db *sql.DB, limit, offset string) ([]Action, error) {
	rows, err := db.Query(`SELECT * FROM actions ORDER BY timestamp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanActions(rows)
}

// FetchActionsByTransactionHash retrieves actions associated with a transaction by its hash
func FetchActionsByTransactionHash(db *sql.DB, txHash string) ([]Action, error) {
	rows, err := db.Query(`SELECT * FROM actions WHERE tx_hash = $1`, txHash)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanActions(rows)
}

// FetchActionsByBlock retrieves actions associated with a block by height or hash
func FetchActionsByBlock(db *sql.DB, blockIdentifier string) ([]Action, error) {
	var query string

	if _, err := strconv.ParseInt(blockIdentifier, 10, 64); err == nil {
		query = `
			SELECT actions.id, actions.tx_hash, actions.action_type, actions.action_name, actions.action_index, actions.input, actions.output, actions.timestamp
			FROM actions
			INNER JOIN transactions ON actions.tx_hash = transactions.tx_hash
			INNER JOIN blocks ON transactions.block_hash = blocks.block_hash
			WHERE blocks.block_height = $1`
	} else {
		query = `
			SELECT actions.id, actions.tx_hash, actions.action_type, actions.action_name, actions.action_index, actions.input, actions.output, actions.timestamp
			FROM actions
			INNER JOIN transactions ON actions.tx_hash = transactions.tx_hash
			WHERE transactions.block_hash = $1`
	}

	rows, err := db.Query(query, blockIdentifier)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanActions(rows)
}

func FetchActionsByType(db *sql.DB, actionType, limit, offset string) ([]Action, error) {
	rows, err := db.Query(`
        SELECT * FROM actions
        WHERE action_type = $1
        ORDER BY timestamp DESC
        LIMIT $2 OFFSET $3`, actionType, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanActions(rows)
}

func FetchActionsByName(db *sql.DB, actionName, limit, offset string) ([]Action, error) {
	rows, err := db.Query(`
        SELECT * FROM actions
        WHERE action_name ILIKE $1
        ORDER BY timestamp DESC
        LIMIT $2 OFFSET $3`, actionName, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanActions(rows)
}

// FetchActionsByUser retrieves actions by user with pagination
func FetchActionsByUser(db *sql.DB, user, limit, offset string) ([]Action, error) {
	normalizedUser := "%" + strings.TrimPrefix(user, "0x") + "%"

	rows, err := db.Query(`
        SELECT actions.*
        FROM actions
        INNER JOIN transactions ON actions.tx_hash = transactions.tx_hash
        WHERE
            transactions.sponsor ILIKE $1
            OR EXISTS (
                SELECT 1
                FROM unnest(transactions.actors) AS actor
                WHERE actor ILIKE $1
            )
            OR EXISTS (
                SELECT 1
                FROM unnest(transactions.receivers) AS receiver
                WHERE receiver ILIKE $1
            )
        ORDER BY actions.timestamp DESC
        LIMIT $2 OFFSET $3
    `, normalizedUser, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanActions(rows)
}

// Helper function to scan action rows and unmarshal action details
func scanActions(rows *sql.Rows) ([]Action, error) {
	var actions []Action

	for rows.Next() {
		var (
			action                            Action
			actionInputJSON, actionOutputJSON []byte
		)
		if err := rows.Scan(&action.ID, &action.TxHash, &action.ActionType, &action.ActionName, &action.ActionIndex, &actionInputJSON, &actionOutputJSON, &action.Timestamp); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(actionInputJSON, &action.Input); err != nil {
			return nil, errors.New("unable to parse action input")
		}
		if err := json.Unmarshal(actionOutputJSON, &action.Output); err != nil {
			return nil, errors.New("unable to parse action output")
		}
		actions = append(actions, action)
	}

	return actions, rows.Err() // Check for errors during iteration
}
