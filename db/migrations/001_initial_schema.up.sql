-- Copyright (C) 2025, Nuklai. All rights reserved.
-- See the file LICENSE for licensing terms.

-- Create schemas
CREATE SCHEMA IF NOT EXISTS blocks;
CREATE SCHEMA IF NOT EXISTS transactions;
CREATE SCHEMA IF NOT EXISTS events;
CREATE SCHEMA IF NOT EXISTS accounts;
CREATE SCHEMA IF NOT EXISTS datasets;
CREATE SCHEMA IF NOT EXISTS validators;
CREATE SCHEMA IF NOT EXISTS governance;
CREATE SCHEMA IF NOT EXISTS wasm;
CREATE SCHEMA IF NOT EXISTS ibc;

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable JSONB operations
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create blocks table
CREATE TABLE IF NOT EXISTS blocks.blocks (
                                             height BIGINT PRIMARY KEY,
                                             hash TEXT NOT NULL,
                                             proposer_address TEXT NOT NULL,
                                             num_txs BIGINT NOT NULL,
                                             total_gas BIGINT NOT NULL,
                                             size BIGINT NOT NULL,
                                             time TIMESTAMP WITH TIME ZONE NOT NULL,
                                             app_hash TEXT NOT NULL,
                                             version BIGINT NOT NULL,
                                             chain_id TEXT NOT NULL,
                                             last_block_id TEXT NOT NULL,
                                             last_commit_hash TEXT NOT NULL,
                                             validators_hash TEXT NOT NULL,
                                             next_validators_hash TEXT NOT NULL,
                                             consensus_hash TEXT NOT NULL,
                                             data_hash TEXT NOT NULL,
                                             evidence_hash TEXT NOT NULL,
                                             created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create index on blocks.time
CREATE INDEX IF NOT EXISTS idx_blocks_time ON blocks.blocks(time);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions.transactions (
                                                         hash TEXT PRIMARY KEY,
                                                         block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    block_time TIMESTAMP WITH TIME ZONE NOT NULL,
    index INTEGER NOT NULL,
    gas_wanted BIGINT NOT NULL,
    gas_used BIGINT NOT NULL,
    fee TEXT NOT NULL,
    memo TEXT NOT NULL,
    error_log TEXT,
    status TEXT NOT NULL,
    codespace TEXT,
    messages JSONB NOT NULL,
    signers JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on transactions
CREATE INDEX IF NOT EXISTS idx_transactions_block_height ON transactions.transactions(block_height);
CREATE INDEX IF NOT EXISTS idx_transactions_block_time ON transactions.transactions(block_time);
CREATE INDEX IF NOT EXISTS idx_transactions_signers_gin ON transactions.transactions USING GIN (signers);

-- Create events table
CREATE TABLE IF NOT EXISTS events.events (
                                             id TEXT PRIMARY KEY,
                                             block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    block_time TIMESTAMP WITH TIME ZONE NOT NULL,
                                                                                                               tx_hash TEXT REFERENCES transactions.transactions(hash) ON DELETE CASCADE,
    tx_index INTEGER,
    event_type TEXT NOT NULL,
    event_index INTEGER NOT NULL,
    module TEXT NOT NULL,
    attributes JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on events
CREATE INDEX IF NOT EXISTS idx_events_block_height ON events.events(block_height);
CREATE INDEX IF NOT EXISTS idx_events_block_time ON events.events(block_time);
CREATE INDEX IF NOT EXISTS idx_events_tx_hash ON events.events(tx_hash);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events.events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_module ON events.events(module);
CREATE INDEX IF NOT EXISTS idx_events_attributes_gin ON events.events USING GIN (attributes);

-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts.accounts (
                                                 address TEXT PRIMARY KEY,
                                                 account_type TEXT NOT NULL,
                                                 balance JSONB NOT NULL,
                                                 last_active TIMESTAMP WITH TIME ZONE NOT NULL,
                                                 created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on accounts
CREATE INDEX IF NOT EXISTS idx_accounts_last_active ON accounts.accounts(last_active);

-- Create datasets table
CREATE TABLE IF NOT EXISTS datasets.datasets (
                                                 denom TEXT PRIMARY KEY,
                                                 name TEXT NOT NULL,
                                                 ticker TEXT NOT NULL,
                                                 description TEXT NOT NULL,
                                                 url TEXT NOT NULL,
                                                 supply TEXT NOT NULL DEFAULT '0',
                                                 categories TEXT NOT NULL,
                                                 license_name TEXT NOT NULL,
                                                 license_symbol TEXT NOT NULL,
                                                 license_url TEXT NOT NULL,
                                                 is_community_dataset BOOLEAN NOT NULL,
                                                 metadata TEXT NOT NULL,
                                                 owner TEXT NOT NULL,
                                                 created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on datasets
CREATE INDEX IF NOT EXISTS idx_datasets_name ON datasets.datasets(name);
CREATE INDEX IF NOT EXISTS idx_datasets_owner ON datasets.datasets(owner);
CREATE INDEX IF NOT EXISTS idx_datasets_categories ON datasets.datasets(categories);

-- Create dataset contributions table
CREATE TABLE IF NOT EXISTS datasets.contributions (
                                                      contribution_id TEXT PRIMARY KEY,
                                                      denom TEXT NOT NULL REFERENCES datasets.datasets(denom) ON DELETE CASCADE,
    data_contributor TEXT NOT NULL,
    data_location TEXT NOT NULL,
    data_identifier TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on dataset contributions
CREATE INDEX IF NOT EXISTS idx_contributions_denom ON datasets.contributions(denom);
CREATE INDEX IF NOT EXISTS idx_contributions_data_contributor ON datasets.contributions(data_contributor);

-- Create pending dataset contributions table
CREATE TABLE IF NOT EXISTS datasets.pending_contributions (
                                                              contribution_id TEXT PRIMARY KEY,
                                                              denom TEXT NOT NULL REFERENCES datasets.datasets(denom) ON DELETE CASCADE,
    data_contributor TEXT NOT NULL,
    data_location TEXT NOT NULL,
    data_identifier TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on pending dataset contributions
CREATE INDEX IF NOT EXISTS idx_pending_contributions_denom ON datasets.pending_contributions(denom);
CREATE INDEX IF NOT EXISTS idx_pending_contributions_data_contributor ON datasets.pending_contributions(data_contributor);

-- Create validators table
CREATE TABLE IF NOT EXISTS validators.validators (
                                                     consensus_address TEXT PRIMARY KEY,
                                                     operator_address TEXT NOT NULL UNIQUE,
                                                     account_address TEXT NOT NULL UNIQUE,
                                                     moniker TEXT NOT NULL,
                                                     identity TEXT,
                                                     website TEXT,
                                                     security_contact TEXT,
                                                     details TEXT,
                                                     commission_rate TEXT NOT NULL,
                                                     voting_power TEXT NOT NULL,
                                                     status TEXT NOT NULL,
                                                     created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on validators
CREATE INDEX IF NOT EXISTS idx_validators_operator_address ON validators.validators(operator_address);
CREATE INDEX IF NOT EXISTS idx_validators_account_address ON validators.validators(account_address);
CREATE INDEX IF NOT EXISTS idx_validators_status ON validators.validators(status);

-- Create validators rewards table
CREATE TABLE IF NOT EXISTS validators.rewards (
    id TEXT PRIMARY KEY,
    validator_address TEXT NOT NULL REFERENCES validators.validators(operator_address),
    delegator_address TEXT NOT NULL,
    amount JSONB NOT NULL,
    block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    block_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes on validators rewards
CREATE INDEX IF NOT EXISTS idx_rewards_validator_address ON validators.rewards(validator_address);
CREATE INDEX IF NOT EXISTS idx_rewards_delegator_address ON validators.rewards(delegator_address);
CREATE INDEX IF NOT EXISTS idx_rewards_block_height ON validators.rewards(block_height);
CREATE INDEX IF NOT EXISTS idx_rewards_block_time ON validators.rewards(block_time);

-- Create validator uptime table
CREATE TABLE IF NOT EXISTS validators.uptime (
                                                 id TEXT PRIMARY KEY,
                                                 consensus_address TEXT NOT NULL REFERENCES validators.validators(consensus_address) ON DELETE CASCADE,
    block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    block_time TIMESTAMP WITH TIME ZONE NOT NULL,
    signed BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(consensus_address, block_height)
    );

-- Create indexes on validator uptime
CREATE INDEX IF NOT EXISTS idx_uptime_consensus_address ON validators.uptime(consensus_address);
CREATE INDEX IF NOT EXISTS idx_uptime_block_height ON validators.uptime(block_height);
CREATE INDEX IF NOT EXISTS idx_uptime_signed ON validators.uptime(signed);

-- Create validators slashing events table
CREATE TABLE IF NOT EXISTS validators.slashing_events (
    id TEXT PRIMARY KEY,
    validator_address TEXT NOT NULL REFERENCES validators.validators(operator_address),
    block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    block_time TIMESTAMP WITH TIME ZONE NOT NULL,
    slash_reason TEXT NOT NULL,
    slash_amount TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes on validators slashing events
CREATE INDEX IF NOT EXISTS idx_slashing_events_validator_address ON validators.slashing_events(validator_address);
CREATE INDEX IF NOT EXISTS idx_slashing_events_block_height ON validators.slashing_events(block_height);
CREATE INDEX IF NOT EXISTS idx_slashing_events_block_time ON validators.slashing_events(block_time);

-- Create delegations table
CREATE TABLE IF NOT EXISTS validators.delegations (
                                                      id TEXT PRIMARY KEY,
                                                      delegator_address TEXT NOT NULL,
                                                      validator_address TEXT NOT NULL REFERENCES validators.validators(operator_address) ON DELETE CASCADE,
    amount TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(delegator_address, validator_address)
    );

-- Create indexes on delegations
CREATE INDEX IF NOT EXISTS idx_delegations_delegator_address ON validators.delegations(delegator_address);
CREATE INDEX IF NOT EXISTS idx_delegations_validator_address ON validators.delegations(validator_address);

-- Create governance proposals table
CREATE TABLE IF NOT EXISTS governance.proposals (
                                                    proposal_id BIGINT PRIMARY KEY,
                                                    title TEXT NOT NULL,
                                                    description TEXT NOT NULL,
                                                    proposal_type TEXT NOT NULL,
                                                    proposal_status TEXT NOT NULL,
                                                    submit_time TIMESTAMP WITH TIME ZONE NOT NULL,
                                                    deposit_end_time TIMESTAMP WITH TIME ZONE NOT NULL,
                                                    voting_start_time TIMESTAMP WITH TIME ZONE,
                                                    voting_end_time TIMESTAMP WITH TIME ZONE,
                                                    proposer TEXT NOT NULL,
                                                    content JSONB NOT NULL,
                                                    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on governance proposals
CREATE INDEX IF NOT EXISTS idx_proposals_proposal_status ON governance.proposals(proposal_status);
CREATE INDEX IF NOT EXISTS idx_proposals_proposer ON governance.proposals(proposer);
CREATE INDEX IF NOT EXISTS idx_proposals_voting_end_time ON governance.proposals(voting_end_time);

-- Create proposal deposits table
CREATE TABLE IF NOT EXISTS governance.deposits (
                                                   id TEXT PRIMARY KEY,
                                                   proposal_id BIGINT NOT NULL REFERENCES governance.proposals(proposal_id) ON DELETE CASCADE,
    depositor_address TEXT NOT NULL,
    amount JSONB NOT NULL,
    block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on proposal deposits
CREATE INDEX IF NOT EXISTS idx_deposits_proposal_id ON governance.deposits(proposal_id);
CREATE INDEX IF NOT EXISTS idx_deposits_depositor_address ON governance.deposits(depositor_address);

-- Create proposal votes table
CREATE TABLE IF NOT EXISTS governance.votes (
                                                id TEXT PRIMARY KEY,
                                                proposal_id BIGINT NOT NULL REFERENCES governance.proposals(proposal_id) ON DELETE CASCADE,
    voter_address TEXT NOT NULL,
    option TEXT NOT NULL,
    block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(proposal_id, voter_address)
    );

-- Create indexes on proposal votes
CREATE INDEX IF NOT EXISTS idx_votes_proposal_id ON governance.votes(proposal_id);
CREATE INDEX IF NOT EXISTS idx_votes_voter_address ON governance.votes(voter_address);
CREATE INDEX IF NOT EXISTS idx_votes_option ON governance.votes(option);

-- Create WASM code table
CREATE TABLE IF NOT EXISTS wasm.codes (
                                          code_id BIGINT PRIMARY KEY,
                                          creator TEXT NOT NULL,
                                          data_hash TEXT NOT NULL,
                                          upload_tx_hash TEXT NOT NULL REFERENCES transactions.transactions(hash) ON DELETE CASCADE,
    instantiated_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on WASM code
CREATE INDEX IF NOT EXISTS idx_codes_creator ON wasm.codes(creator);

-- Create WASM contracts table
CREATE TABLE IF NOT EXISTS wasm.contracts (
                                              address TEXT PRIMARY KEY,
                                              code_id BIGINT NOT NULL REFERENCES wasm.codes(code_id) ON DELETE CASCADE,
    creator TEXT NOT NULL,
    admin TEXT,
    label TEXT NOT NULL,
    created_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    instantiate_tx_hash TEXT NOT NULL REFERENCES transactions.transactions(hash) ON DELETE CASCADE,
    init_msg JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on WASM contracts
CREATE INDEX IF NOT EXISTS idx_contracts_code_id ON wasm.contracts(code_id);
CREATE INDEX IF NOT EXISTS idx_contracts_creator ON wasm.contracts(creator);
CREATE INDEX IF NOT EXISTS idx_contracts_admin ON wasm.contracts(admin);
CREATE INDEX IF NOT EXISTS idx_contracts_label ON wasm.contracts(label);

-- Create WASM executions table
CREATE TABLE IF NOT EXISTS wasm.executions (
                                               id TEXT PRIMARY KEY,
                                               contract_address TEXT NOT NULL REFERENCES wasm.contracts(address) ON DELETE CASCADE,
    sender TEXT NOT NULL,
    tx_hash TEXT NOT NULL REFERENCES transactions.transactions(hash) ON DELETE CASCADE,
    block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    block_time TIMESTAMP WITH TIME ZONE NOT NULL,
    execute_msg JSONB NOT NULL,
    funds JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on WASM executions
CREATE INDEX IF NOT EXISTS idx_executions_contract_address ON wasm.executions(contract_address);
CREATE INDEX IF NOT EXISTS idx_executions_sender ON wasm.executions(sender);
CREATE INDEX IF NOT EXISTS idx_executions_block_height ON wasm.executions(block_height);
CREATE INDEX IF NOT EXISTS idx_executions_block_time ON wasm.executions(block_time);

-- Create WASM queries table
CREATE TABLE IF NOT EXISTS wasm.queries (
                                            id TEXT PRIMARY KEY,
                                            contract_address TEXT NOT NULL REFERENCES wasm.contracts(address) ON DELETE CASCADE,
    sender TEXT NOT NULL,
    query_msg JSONB NOT NULL,
    response JSONB NOT NULL,
    block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on WASM queries
CREATE INDEX IF NOT EXISTS idx_queries_contract_address ON wasm.queries(contract_address);
CREATE INDEX IF NOT EXISTS idx_queries_sender ON wasm.queries(sender);
CREATE INDEX IF NOT EXISTS idx_queries_block_height ON wasm.queries(block_height);

-- Create IBC channels table
CREATE TABLE IF NOT EXISTS ibc.channels (
                                            channel_id TEXT NOT NULL,
                                            port_id TEXT NOT NULL,
                                            connection_id TEXT NOT NULL,
                                            counterparty_channel_id TEXT NOT NULL,
                                            counterparty_port_id TEXT NOT NULL,
                                            state TEXT NOT NULL,
                                            ordering TEXT NOT NULL,
                                            version TEXT NOT NULL,
                                            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, port_id)
    );

-- Create indexes on IBC channels
CREATE INDEX IF NOT EXISTS idx_channels_connection_id ON ibc.channels(connection_id);
CREATE INDEX IF NOT EXISTS idx_channels_state ON ibc.channels(state);

-- Create IBC connections table
CREATE TABLE IF NOT EXISTS ibc.connections (
                                               connection_id TEXT PRIMARY KEY,
                                               client_id TEXT NOT NULL,
                                               counterparty_connection_id TEXT NOT NULL,
                                               counterparty_client_id TEXT NOT NULL,
                                               state TEXT NOT NULL,
                                               created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on IBC connections
CREATE INDEX IF NOT EXISTS idx_connections_client_id ON ibc.connections(client_id);
CREATE INDEX IF NOT EXISTS idx_connections_state ON ibc.connections(state);

-- Create IBC clients table
CREATE TABLE IF NOT EXISTS ibc.clients (
                                           client_id TEXT PRIMARY KEY,
                                           client_type TEXT NOT NULL,
                                           chain_id TEXT NOT NULL,
                                           last_height BIGINT NOT NULL,
                                           state TEXT NOT NULL,
                                           created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Create indexes on IBC clients
CREATE INDEX IF NOT EXISTS idx_clients_client_type ON ibc.clients(client_type);
CREATE INDEX IF NOT EXISTS idx_clients_chain_id ON ibc.clients(chain_id);
CREATE INDEX IF NOT EXISTS idx_clients_state ON ibc.clients(state);

-- Create IBC packets table
CREATE TABLE IF NOT EXISTS ibc.packets (
                                           id TEXT PRIMARY KEY,
                                           source_channel_id TEXT NOT NULL,
                                           source_port_id TEXT NOT NULL,
                                           dest_channel_id TEXT NOT NULL,
                                           dest_port_id TEXT NOT NULL,
                                           sequence_number BIGINT NOT NULL,
                                           packet_type TEXT NOT NULL,
                                           packet_status TEXT NOT NULL,
                                           sent_block_height BIGINT NOT NULL REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    sent_block_time TIMESTAMP WITH TIME ZONE NOT NULL,
                                                                                                                  received_block_height BIGINT REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    received_block_time TIMESTAMP WITH TIME ZONE,
                                                                                                                  ack_block_height BIGINT REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    ack_block_time TIMESTAMP WITH TIME ZONE,
                                                                                                                  timeout_block_height BIGINT REFERENCES blocks.blocks(height) ON DELETE CASCADE,
    timeout_block_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(source_channel_id, source_port_id, sequence_number)
    );

-- Create indexes on IBC packets
CREATE INDEX IF NOT EXISTS idx_packets_source_channel_id_port_id ON ibc.packets(source_channel_id, source_port_id);
CREATE INDEX IF NOT EXISTS idx_packets_dest_channel_id_port_id ON ibc.packets(dest_channel_id, dest_port_id);
CREATE INDEX IF NOT EXISTS idx_packets_packet_status ON ibc.packets(packet_status);
