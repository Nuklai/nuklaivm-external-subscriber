#!/bin/bash

# Parse command line arguments for nuking data
NUKE=false
while getopts "n" opt; do
  case ${opt} in
    n )
      NUKE=true
      ;;
    * )
      echo "Usage: $0 [-n (nuke database)]"
      exit 1
      ;;
  esac
done

rm -f block_subscriber.log

if [ "$NUKE" = true ]; then
  # Stop and remove old Docker container if running
  echo "Stopping and removing old PostgreSQL container if exists..."
  docker-compose down -v --remove-orphans
else
  # Stop and remove old Docker container if running, keeping volumes intact
  echo "Stopping old PostgreSQL container if exists..."
  docker-compose down
fi

# Kill any process using port 50051 (gRPC server port)
echo "Killing any process using port 50051..."
fuser -k 50051/tcp

# Kill any process using port 8080 (REST API server port)
echo "Killing any process using port 8080..."
fuser -k 8080/tcp

# Start Docker container
echo "Starting PostgreSQL with TimescaleDB..."
docker-compose up -d

# Wait for the database to be ready
echo "Waiting for PostgreSQL to start..."
sleep 10

# Create database schema
echo "Creating database schema..."
docker exec -i timescaledb psql -U postgres -d blockchain << EOF
-- Block Table
CREATE TABLE IF NOT EXISTS blocks (
    id SERIAL PRIMARY KEY,
    block_height BIGINT UNIQUE NOT NULL,
    block_hash TEXT NOT NULL,
    parent_block_hash TEXT,
    state_root TEXT,
    timestamp TIMESTAMPTZ NOT NULL,
    unit_prices TEXT
);

-- Transaction Table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    tx_hash TEXT UNIQUE NOT NULL,
    block_hash TEXT,
    sponsor TEXT,
    max_fee NUMERIC,
    success BOOLEAN,
    fee NUMERIC,
    outputs JSONB,
    timestamp TIMESTAMPTZ NOT NULL
);

-- Actions Table
CREATE TABLE IF NOT EXISTS actions (
    id SERIAL PRIMARY KEY,
    tx_hash TEXT UNIQUE NOT NULL,
    action_type SMALLINT NOT NULL,
    action_details JSONB,
    timestamp TIMESTAMPTZ NOT NULL
);

-- Genesis Data Table without unique constraint
CREATE TABLE IF NOT EXISTS genesis_data (
    id SERIAL PRIMARY KEY,
    data JSONB
);

-- Non-unique indexes for optimized querying
CREATE INDEX IF NOT EXISTS idx_block_height ON blocks(block_height);
CREATE INDEX IF NOT EXISTS idx_block_hash ON blocks(block_hash);
CREATE INDEX IF NOT EXISTS idx_tx_hash ON transactions(tx_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_block_hash ON transactions(block_hash);
CREATE INDEX IF NOT EXISTS idx_sponsor ON transactions(sponsor);
CREATE INDEX IF NOT EXISTS idx_action_type ON actions(action_type);

-- Optional JSONB indexes for performance with JSONB data
CREATE INDEX IF NOT EXISTS idx_outputs_jsonb ON transactions USING GIN (outputs);
CREATE INDEX IF NOT EXISTS idx_action_details_jsonb ON actions USING GIN (action_details);
EOF

# Run the Go block subscriber server
echo "Starting the Go block subscriber server..."
nohup go run main.go &> block_subscriber.log &

# Sleep to prevent race conditions with duplicate instances
sleep 5

# Notify that everything is set up
echo "Setup complete. The Go block subscriber server is running."
