// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/nuklai/nuklaivm-external-subscriber/config"
)

// InitDB initializes the database connection and creates the schema if it doesn't exist
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging the database: %w", err)
	}

	log.Println("Database connection established")

	reset := config.GetEnv("DB_RESET", "false") == "true"
	if reset {
		// Drop all existing tables
		log.Println("Resetting the database...")
		_, err := db.Exec(`
			DROP TABLE IF EXISTS blocks, transactions, actions, assets, genesis_data CASCADE;
		`)
		if err != nil {
			return nil, fmt.Errorf("error resetting the database: %w", err)
		}
	}

	// Ensure schema is created
	if err := CreateSchema(db); err != nil {
		return nil, fmt.Errorf("error creating schema: %w", err)
	}

	return db, nil
}

// CreateSchema creates the database schema if it doesn't already exist
func CreateSchema(db *sql.DB) error {
	schema := `
	-- Ensure the pg_trgm extension is enabled for GIN indexes
	CREATE EXTENSION IF NOT EXISTS pg_trgm;

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
		actors TEXT[],
		receivers TEXT[],
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

	CREATE TABLE IF NOT EXISTS validator_stake (
    id SERIAL PRIMARY KEY,
    node_id TEXT NOT NULL UNIQUE,
    actor TEXT NOT NULL,
    stake_start_block BIGINT NOT NULL,
    stake_end_block BIGINT NOT NULL,
    staked_amount BIGINT NOT NULL,
    delegation_fee_rate BIGINT NOT NULL,
    reward_address TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    UNIQUE (node_id, stake_start_block)
	);

	CREATE TABLE IF NOT EXISTS health_events (
    id SERIAL PRIMARY KEY,
    state VARCHAR(10) NOT NULL,
    description TEXT NOT NULL,
    service_names TEXT[],
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration INT,
    timestamp TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS daily_health_summaries (
    date DATE PRIMARY KEY,
    state VARCHAR(10) NOT NULL,
    incidents TEXT[],
    last_updated TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS action_volumes (
    action_type SMALLINT PRIMARY KEY,
    action_name TEXT NOT NULL,
    total_count BIGINT NOT NULL DEFAULT 0
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
	CREATE INDEX IF NOT EXISTS idx_actors ON transactions USING GIN (actors);
	CREATE INDEX IF NOT EXISTS idx_receivers ON transactions USING GIN (receivers);

	CREATE INDEX IF NOT EXISTS idx_action_type ON actions(action_type);
	CREATE INDEX IF NOT EXISTS idx_action_name_lower ON actions (LOWER(action_name));

	CREATE INDEX IF NOT EXISTS idx_assets_creator ON assets(asset_creator);
  	CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(asset_type_id);
	CREATE INDEX IF NOT EXISTS idx_asset_address ON assets(asset_address);

	CREATE INDEX IF NOT EXISTS idx_validator_stake_node_id ON validator_stake(node_id);
	CREATE INDEX IF NOT EXISTS idx_validator_stake_actor ON validator_stake(actor);
	CREATE INDEX IF NOT EXISTS idx_validator_stake_timestamp ON validator_stake(timestamp);
	CREATE INDEX IF NOT EXISTS idx_health_events_state ON health_events(state);
	CREATE INDEX IF NOT EXISTS idx_health_events_timestamp ON health_events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_daily_health_summaries_date ON daily_health_summaries(date);
	CREATE INDEX IF NOT EXISTS idx_daily_health_summaries_state ON daily_health_summaries(state);
	CREATE INDEX IF NOT EXISTS idx_daily_health_summaries_last_updated ON daily_health_summaries(last_updated);
	CREATE INDEX IF NOT EXISTS idx_action_volumes_name ON action_volumes(action_name);

	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("error executing schema creation: %w", err)
	}

	_, err = db.Exec(`
    DO $$ 
    BEGIN 
        IF NOT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_name = 'health_events' 
            AND column_name = 'service_names'
        ) THEN
            ALTER TABLE health_events 
            ADD COLUMN service_names TEXT[];
        END IF;
    END $$;
    `)
	if err != nil {
		return fmt.Errorf("error service_names column: %w", err)
	}

	log.Println("Database schema created or already exists")
	return nil
}
