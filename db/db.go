// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// InitDB initializes the database connection and creates the schema if it doesn't exist
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging the database: %w", err)
	}

	log.Println("Database connection established")

	// Ensure schema is created
	if err := CreateSchema(db); err != nil {
		return nil, fmt.Errorf("error creating schema: %w", err)
	}

	return db, nil
}

// CreateSchema creates the database schema if it doesn't already exist
func CreateSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS blocks (
			block_height BIGINT PRIMARY KEY,
			block_hash TEXT NOT NULL,
			parent_block_hash TEXT,
			state_root TEXT,
			block_size INT NOT NULL,
			tx_count INT NOT NULL,
			total_fee NUMERIC NOT NULL,
			avg_tx_size NUMERIC NOT NULL,
			unique_participants INT NOT NULL,
			timestamp TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		tx_hash TEXT UNIQUE NOT NULL,
		block_hash TEXT,
		sponsor TEXT,
		max_fee NUMERIC,
		success BOOLEAN,
		fee NUMERIC,
		actions JSON,
		timestamp TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS actions (
		id SERIAL PRIMARY KEY,
		tx_hash TEXT NOT NULL,
		action_type SMALLINT NOT NULL,
		action_name TEXT,
		action_index INT NOT NULL,
		input JSON,
    output JSON,
		timestamp TIMESTAMP NOT NULL,
		UNIQUE (tx_hash, action_type, action_index)
	);

  CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY,
    asset_address TEXT NOT NULL UNIQUE,
    asset_type_id SMALLINT NOT NULL,
    asset_type TEXT NOT NULL,
    asset_creator TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    name TEXT,
    symbol TEXT,
    decimals INT,
    metadata TEXT,
    max_supply NUMERIC,
    mint_admin TEXT,
    pause_unpause_admin TEXT,
    freeze_unfreeze_admin TEXT,
    enable_disable_kyc_account_admin TEXT,
    timestamp TIMESTAMP NOT NULL
);

	CREATE TABLE IF NOT EXISTS genesis_data (
		id SERIAL PRIMARY KEY,
		data JSON
	);

	CREATE INDEX IF NOT EXISTS idx_block_height ON blocks(block_height);
	CREATE INDEX IF NOT EXISTS idx_block_hash ON blocks(block_hash);

	CREATE INDEX IF NOT EXISTS idx_tx_hash ON transactions(tx_hash);
	CREATE INDEX IF NOT EXISTS idx_transactions_block_hash ON transactions(block_hash);
	CREATE INDEX IF NOT EXISTS idx_sponsor ON transactions(sponsor);
	CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions (timestamp);

	CREATE INDEX IF NOT EXISTS idx_action_type ON actions(action_type);
	CREATE INDEX IF NOT EXISTS idx_action_name_lower ON actions (LOWER(action_name));

	CREATE INDEX IF NOT EXISTS idx_assets_creator ON assets(asset_creator);
  CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(asset_type_id);
	CREATE INDEX IF NOT EXISTS idx_asset_address ON assets(asset_address);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("error executing schema creation: %w", err)
	}

	log.Println("Database schema created or already exists")
	return nil
}
