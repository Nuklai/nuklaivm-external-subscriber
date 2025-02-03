// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql"
	"fmt"
	"log"
)

type AccountStats struct {
	TotalAccounts  int `json:"total_accounts"`
	TotalNAIHeld   int `json:"total_nai_held"`
	ActiveAccounts int `json:"active_accounts"`
}

type Account struct {
	Address          string  `json:"address"`
	Balance          float64 `json:"balance"`
	TransactionCount int     `json:"transaction_count"`
}

// FetchAccountStats retrieves all account stats
func FetchAccountStats(db *sql.DB) (AccountStats, error) {
	var stats AccountStats

	// Get the total unique addresses from actors & receivers from transactions table
	err := db.QueryRow(`
        WITH unique_addresses AS (
            SELECT DISTINCT sponsor FROM transactions
            UNION
            SELECT DISTINCT UNNEST(actors) FROM transactions
            UNION
            SELECT DISTINCT UNNEST(receivers) FROM transactions
        )
        SELECT COUNT(*) FROM unique_addresses
    `).Scan(&stats.TotalAccounts)
	if err != nil {
		log.Printf("Error counting total accounts: %v", err)
		return stats, err
	}

	err = db.QueryRow(`
        WITH latest_balances AS (
            SELECT 
                CASE 
                    WHEN TRIM(BOTH '"' FROM REPLACE(t.sponsor, '0x', '')) = TRIM(BOTH '"' FROM REPLACE(a.input->>'to', '0x', ''))
                        THEN CAST(a.output->>'receiver_balance' AS NUMERIC)
                    ELSE CAST(a.output->>'sender_balance' AS NUMERIC)
                END as balance,
                t.sponsor as address
            FROM transactions t
            JOIN actions a ON t.tx_hash = a.tx_hash
            WHERE a.action_type = 0 
            AND a.action_name = 'Transfer'
            UNION ALL
            SELECT 
                CAST(a.output->>'receiver_balance' AS NUMERIC) as balance,
                r.receiver as address
            FROM transactions t
            CROSS JOIN UNNEST(t.receivers) as r(receiver)
            JOIN actions a ON t.tx_hash = a.tx_hash
            WHERE a.action_type = 0 
            AND a.action_name = 'Transfer'
        )
        SELECT COALESCE(SUM(balance), 0)
        FROM (
            SELECT DISTINCT ON (address) address, balance
            FROM latest_balances
            ORDER BY address, balance DESC
        ) final_balances
    `).Scan(&stats.TotalNAIHeld)
	if err != nil {
		log.Printf("Error calculating total NAI held: %v", err)
		return stats, err
	}

	// Get active accounts - 24h
	err = db.QueryRow(`
        WITH active_addresses AS (
            SELECT DISTINCT sponsor FROM transactions 
            WHERE timestamp >= NOW() - INTERVAL '24 hours'
            UNION
            SELECT DISTINCT UNNEST(actors) FROM transactions
            WHERE timestamp >= NOW() - INTERVAL '24 hours'
            UNION
            SELECT DISTINCT UNNEST(receivers) FROM transactions
            WHERE timestamp >= NOW() - INTERVAL '24 hours'
        )
        SELECT COUNT(*) FROM active_addresses
    `).Scan(&stats.ActiveAccounts)
	if err != nil {
		log.Printf("Error counting active accounts: %v", err)
		return stats, err
	}

	return stats, nil
}

func CountAccounts(db *sql.DB) (int, error) {
	var count int
	query := `
        WITH unique_addresses AS (
            SELECT DISTINCT sponsor FROM transactions
            UNION
            SELECT DISTINCT UNNEST(actors) FROM transactions
            UNION
            SELECT DISTINCT UNNEST(receivers) FROM transactions
        )
        SELECT COUNT(*) FROM unique_addresses
    `

	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("Error counting accounts: %v", err)
		return 0, err
	}

	return count, nil
}

// FetchAllAccounts retrieves accounts address, balance, transaction count
func FetchAllAccounts(db *sql.DB, limit, offset string) ([]Account, error) {
	query := `
        WITH unique_addresses AS (
            SELECT DISTINCT sponsor as address FROM transactions
            UNION
            SELECT DISTINCT UNNEST(actors) as address FROM transactions
            UNION
            SELECT DISTINCT UNNEST(receivers) as address FROM transactions
        ), transfer_balances AS (
            SELECT 
                CASE 
                    WHEN TRIM(BOTH '"' FROM REPLACE(t.sponsor, '0x', '')) = TRIM(BOTH '"' FROM REPLACE(a.input->>'to', '0x', ''))
                        THEN CAST(a.output->>'receiver_balance' AS NUMERIC)
                    ELSE CAST(a.output->>'sender_balance' AS NUMERIC)
                END as balance,
                t.sponsor as address
            FROM transactions t
            JOIN actions a ON t.tx_hash = a.tx_hash
            WHERE a.action_type = 0 
            AND a.action_name = 'Transfer'
            UNION ALL
            SELECT 
                CAST(a.output->>'receiver_balance' AS NUMERIC) as balance,
                r.receiver as address
            FROM transactions t
            CROSS JOIN UNNEST(t.receivers) as r(receiver)
            JOIN actions a ON t.tx_hash = a.tx_hash
            WHERE a.action_type = 0 
            AND a.action_name = 'Transfer'
        ), latest_balances AS (
            SELECT DISTINCT ON (address) 
                address,
                balance
            FROM transfer_balances
            ORDER BY address, balance DESC
        )
        SELECT 
            ua.address,
            COALESCE(lb.balance, 0) as balance,
            COUNT(DISTINCT t.tx_hash) as tx_count
        FROM unique_addresses ua
        LEFT JOIN latest_balances lb ON ua.address = lb.address
        LEFT JOIN transactions t ON 
            ua.address = t.sponsor 
            OR ua.address = ANY(t.actors) 
            OR ua.address = ANY(t.receivers)
        GROUP BY ua.address, lb.balance
        ORDER BY COUNT(DISTINCT t.tx_hash) DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.Address, &account.Balance, &account.TransactionCount); err != nil {
			log.Printf("Error scanning account row: %v", err)
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating account rows: %v", err)
		return nil, err
	}

	return accounts, nil
}

// FetchAccountByAddress retrieves details for a specific account/address
func FetchAccountByAddress(db *sql.DB, address string) (*Account, error) {
	query := `
        WITH unique_addresses AS (
            SELECT DISTINCT sponsor as address FROM transactions
            UNION
            SELECT DISTINCT UNNEST(actors) as address FROM transactions
            UNION
            SELECT DISTINCT UNNEST(receivers) as address FROM transactions
        ), transfer_balances AS (
            SELECT 
                CASE 
                    WHEN TRIM(BOTH '"' FROM REPLACE(t.sponsor, '0x', '')) = TRIM(BOTH '"' FROM REPLACE(a.input->>'to', '0x', ''))
                        THEN CAST(a.output->>'receiver_balance' AS NUMERIC)
                    ELSE CAST(a.output->>'sender_balance' AS NUMERIC)
                END as balance,
                t.sponsor as address
            FROM transactions t
            JOIN actions a ON t.tx_hash = a.tx_hash
            WHERE a.action_type = 0 
            AND a.action_name = 'Transfer'
            UNION ALL
            SELECT 
                CAST(a.output->>'receiver_balance' AS NUMERIC) as balance,
                r.receiver as address
            FROM transactions t
            CROSS JOIN UNNEST(t.receivers) as r(receiver)
            JOIN actions a ON t.tx_hash = a.tx_hash
            WHERE a.action_type = 0 
            AND a.action_name = 'Transfer'
        ), latest_balances AS (
            SELECT DISTINCT ON (address) 
                address,
                balance
            FROM transfer_balances
            ORDER BY address, balance DESC
        )
        SELECT 
            ua.address,
            COALESCE(lb.balance, 0) as balance,
            COUNT(DISTINCT t.tx_hash) as tx_count
        FROM unique_addresses ua
        LEFT JOIN latest_balances lb ON ua.address = lb.address
        LEFT JOIN transactions t ON 
            ua.address = t.sponsor 
            OR ua.address = ANY(t.actors) 
            OR ua.address = ANY(t.receivers)
        WHERE ua.address = $1
        GROUP BY ua.address, lb.balance
    `

	row := db.QueryRow(query, address)

	var account Account
	if err := row.Scan(&account.Address, &account.Balance, &account.TransactionCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no account found for address %s", address)
		}
		log.Printf("Error scanning account row: %v", err)
		return nil, err
	}

	return &account, nil
}
