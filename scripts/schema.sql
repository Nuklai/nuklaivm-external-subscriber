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