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
	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("error creating schema: %w", err)
	}

	return db, nil
}

// createSchema creates the database schema if it doesn't already exist
func createSchema(db *sql.DB) error {
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
		outputs JSON,
		timestamp TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS actions (
		id SERIAL PRIMARY KEY,
		tx_hash TEXT NOT NULL,
		action_type SMALLINT NOT NULL,
		action_details JSON,
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
	CREATE INDEX IF NOT EXISTS idx_action_type ON actions(action_type);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("error executing schema creation: %w", err)
	}

	log.Println("Database schema created or already exists")
	return nil
}
