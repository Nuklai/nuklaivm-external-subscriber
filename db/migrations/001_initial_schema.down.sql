-- Copyright (C) 2025, Nuklai. All rights reserved.
-- See the file LICENSE for licensing terms.


-- Drop IBC tables
DROP TABLE IF EXISTS ibc.packets;
DROP TABLE IF EXISTS ibc.clients;
DROP TABLE IF EXISTS ibc.connections;
DROP TABLE IF EXISTS ibc.channels;

-- Drop WASM tables
DROP TABLE IF EXISTS wasm.queries;
DROP TABLE IF EXISTS wasm.executions;
DROP TABLE IF EXISTS wasm.contracts;
DROP TABLE IF EXISTS wasm.codes;

-- Drop governance tables
DROP TABLE IF EXISTS governance.votes;
DROP TABLE IF EXISTS governance.deposits;
DROP TABLE IF EXISTS governance.proposals;

-- Drop validator tables
DROP TABLE IF EXISTS validators.delegations;
DROP TABLE IF EXISTS validators.uptime;
DROP TABLE IF EXISTS validators.validators;

-- Drop dataset tables
DROP TABLE IF EXISTS datasets.pending_contributions;
DROP TABLE IF EXISTS datasets.contributions;
DROP TABLE IF EXISTS datasets.datasets;

-- Drop account tables
DROP TABLE IF EXISTS accounts.accounts;

-- Drop event tables
DROP TABLE IF EXISTS events.events;

-- Drop transaction tables
DROP TABLE IF EXISTS transactions.transactions;

-- Drop block tables
DROP TABLE IF EXISTS blocks.blocks;

-- Drop schemas
DROP SCHEMA IF EXISTS ibc;
DROP SCHEMA IF EXISTS wasm;
DROP SCHEMA IF EXISTS governance;
DROP SCHEMA IF EXISTS validators;
DROP SCHEMA IF EXISTS datasets;
DROP SCHEMA IF EXISTS accounts;
DROP SCHEMA IF EXISTS events;
DROP SCHEMA IF EXISTS transactions;
DROP SCHEMA IF EXISTS blocks;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS "pgcrypto";