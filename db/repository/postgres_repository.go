// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
	"github.com/nuklai/nuklaivm-external-subscriber/db/postgres"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *postgres.Database) Repository {
	return &PostgresRepository{
		db: db.Pool(),
	}
}

// Block operations
func (r *PostgresRepository) SaveBlock(ctx context.Context, block *models.Block) error {
	query := `
		INSERT INTO blocks.blocks (
			height, hash, proposer_address, num_txs, total_gas, size, time, 
			app_hash, version, chain_id, last_block_id, last_commit_hash, 
			validators_hash, next_validators_hash, consensus_hash, data_hash, 
			evidence_hash, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (height) DO UPDATE SET
			hash = EXCLUDED.hash,
			proposer_address = EXCLUDED.proposer_address,
			num_txs = EXCLUDED.num_txs,
			total_gas = EXCLUDED.total_gas,
			size = EXCLUDED.size,
			time = EXCLUDED.time,
			app_hash = EXCLUDED.app_hash,
			version = EXCLUDED.version,
			chain_id = EXCLUDED.chain_id,
			last_block_id = EXCLUDED.last_block_id,
			last_commit_hash = EXCLUDED.last_commit_hash,
			validators_hash = EXCLUDED.validators_hash,
			next_validators_hash = EXCLUDED.next_validators_hash,
			consensus_hash = EXCLUDED.consensus_hash,
			data_hash = EXCLUDED.data_hash,
			evidence_hash = EXCLUDED.evidence_hash
	`

	_, err := r.db.Exec(ctx, query,
		block.Height,
		block.Hash,
		block.ProposerAddress,
		block.NumTxs,
		block.TotalGas,
		block.Size,
		block.Time,
		block.AppHash,
		block.Version,
		block.ChainID,
		block.LastBlockID,
		block.LastCommitHash,
		block.ValidatorsHash,
		block.NextValidatorsHash,
		block.ConsensusHash,
		block.DataHash,
		block.EvidenceHash,
		block.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save block at height %d: %w", block.Height, err)
	}

	return nil
}

func (r *PostgresRepository) GetBlockByHeight(ctx context.Context, height int64) (*models.Block, error) {
	query := `
		SELECT 
			height, hash, proposer_address, num_txs, total_gas, size, time,
			app_hash, version, chain_id, last_block_id, last_commit_hash,
			validators_hash, next_validators_hash, consensus_hash, data_hash,
			evidence_hash, created_at
		FROM blocks.blocks
		WHERE height = $1
	`

	var block models.Block
	err := r.db.QueryRow(ctx, query, height).Scan(
		&block.Height,
		&block.Hash,
		&block.ProposerAddress,
		&block.NumTxs,
		&block.TotalGas,
		&block.Size,
		&block.Time,
		&block.AppHash,
		&block.Version,
		&block.ChainID,
		&block.LastBlockID,
		&block.LastCommitHash,
		&block.ValidatorsHash,
		&block.NextValidatorsHash,
		&block.ConsensusHash,
		&block.DataHash,
		&block.EvidenceHash,
		&block.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
	}

	return &block, nil
}

func (r *PostgresRepository) GetBlockByHash(ctx context.Context, hash string) (*models.Block, error) {
	query := `
        SELECT 
            height, hash, proposer_address, num_txs, total_gas, size, time,
            app_hash, version, chain_id, last_block_id, last_commit_hash,
            validators_hash, next_validators_hash, consensus_hash, data_hash,
            evidence_hash, created_at
        FROM blocks.blocks
        WHERE hash = $1
    `

	var block models.Block
	err := r.db.QueryRow(ctx, query, hash).Scan(
		&block.Height,
		&block.Hash,
		&block.ProposerAddress,
		&block.NumTxs,
		&block.TotalGas,
		&block.Size,
		&block.Time,
		&block.AppHash,
		&block.Version,
		&block.ChainID,
		&block.LastBlockID,
		&block.LastCommitHash,
		&block.ValidatorsHash,
		&block.NextValidatorsHash,
		&block.ConsensusHash,
		&block.DataHash,
		&block.EvidenceHash,
		&block.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get block by hash %s: %w", hash, err)
	}

	return &block, nil
}

func (r *PostgresRepository) GetLatestBlock(ctx context.Context) (*models.Block, error) {
	query := `
		SELECT 
			height, hash, proposer_address, num_txs, total_gas, size, time,
			app_hash, version, chain_id, last_block_id, last_commit_hash,
			validators_hash, next_validators_hash, consensus_hash, data_hash,
			evidence_hash, created_at
		FROM blocks.blocks
		ORDER BY height DESC
		LIMIT 1
	`

	var block models.Block
	err := r.db.QueryRow(ctx, query).Scan(
		&block.Height,
		&block.Hash,
		&block.ProposerAddress,
		&block.NumTxs,
		&block.TotalGas,
		&block.Size,
		&block.Time,
		&block.AppHash,
		&block.Version,
		&block.ChainID,
		&block.LastBlockID,
		&block.LastCommitHash,
		&block.ValidatorsHash,
		&block.NextValidatorsHash,
		&block.ConsensusHash,
		&block.DataHash,
		&block.EvidenceHash,
		&block.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return &block, nil
}

func (r *PostgresRepository) GetBlocksInRange(ctx context.Context, startHeight, endHeight int64) ([]*models.Block, error) {
	query := `
        SELECT 
            height, hash, proposer_address, num_txs, total_gas, size, time,
            app_hash, version, chain_id, last_block_id, last_commit_hash,
            validators_hash, next_validators_hash, consensus_hash, data_hash,
            evidence_hash, created_at
        FROM blocks.blocks
        WHERE height >= $1 AND height <= $2
        ORDER BY height ASC
    `

	rows, err := r.db.Query(ctx, query, startHeight, endHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks in range %d-%d: %w", startHeight, endHeight, err)
	}
	defer rows.Close()

	var blocks []*models.Block
	for rows.Next() {
		var block models.Block
		err := rows.Scan(
			&block.Height,
			&block.Hash,
			&block.ProposerAddress,
			&block.NumTxs,
			&block.TotalGas,
			&block.Size,
			&block.Time,
			&block.AppHash,
			&block.Version,
			&block.ChainID,
			&block.LastBlockID,
			&block.LastCommitHash,
			&block.ValidatorsHash,
			&block.NextValidatorsHash,
			&block.ConsensusHash,
			&block.DataHash,
			&block.EvidenceHash,
			&block.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block row: %w", err)
		}
		blocks = append(blocks, &block)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating block rows: %w", err)
	}

	return blocks, nil
}

func (r *PostgresRepository) GetBlocksByTime(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.Block, error) {
	query := `
        SELECT 
            height, hash, proposer_address, num_txs, total_gas, size, time,
            app_hash, version, chain_id, last_block_id, last_commit_hash,
            validators_hash, next_validators_hash, consensus_hash, data_hash,
            evidence_hash, created_at
        FROM blocks.blocks
        WHERE time >= $1 AND time <= $2
        ORDER BY height DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.db.Query(ctx, query, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks by time range: %w", err)
	}
	defer rows.Close()

	var blocks []*models.Block
	for rows.Next() {
		var block models.Block
		err := rows.Scan(
			&block.Height,
			&block.Hash,
			&block.ProposerAddress,
			&block.NumTxs,
			&block.TotalGas,
			&block.Size,
			&block.Time,
			&block.AppHash,
			&block.Version,
			&block.ChainID,
			&block.LastBlockID,
			&block.LastCommitHash,
			&block.ValidatorsHash,
			&block.NextValidatorsHash,
			&block.ConsensusHash,
			&block.DataHash,
			&block.EvidenceHash,
			&block.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block row: %w", err)
		}
		blocks = append(blocks, &block)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating block rows: %w", err)
	}

	return blocks, nil
}

// Txn operations

func (r *PostgresRepository) SaveTransaction(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions.transactions (
			hash, block_height, block_time, index, gas_wanted, gas_used,
			fee, memo, error_log, status, codespace, messages, signers, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (hash) DO UPDATE SET
			block_height = EXCLUDED.block_height,
			block_time = EXCLUDED.block_time,
			index = EXCLUDED.index,
			gas_wanted = EXCLUDED.gas_wanted,
			gas_used = EXCLUDED.gas_used,
			fee = EXCLUDED.fee,
			memo = EXCLUDED.memo,
			error_log = EXCLUDED.error_log,
			status = EXCLUDED.status,
			codespace = EXCLUDED.codespace,
			messages = EXCLUDED.messages,
			signers = EXCLUDED.signers
	`

	messagesJSON, err := json.Marshal(tx.Messages)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction messages: %w", err)
	}

	signersJSON, err := json.Marshal(tx.Signers)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction signers: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		tx.Hash,
		tx.BlockHeight,
		tx.BlockTime,
		tx.Index,
		tx.GasWanted,
		tx.GasUsed,
		tx.Fee,
		tx.Memo,
		tx.ErrorLog,
		tx.Status,
		tx.Codespace,
		messagesJSON,
		signersJSON,
		tx.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save transaction %s: %w", tx.Hash, err)
	}

	return nil
}

func (r *PostgresRepository) GetTransactionByHash(ctx context.Context, hash string) (*models.Transaction, error) {
	query := `
        SELECT 
            hash, block_height, block_time, index, gas_wanted, gas_used,
            fee, memo, error_log, status, codespace, messages, signers, created_at
        FROM transactions.transactions
        WHERE hash = $1
    `

	var tx models.Transaction
	var messagesJSON, signersJSON []byte

	err := r.db.QueryRow(ctx, query, hash).Scan(
		&tx.Hash,
		&tx.BlockHeight,
		&tx.BlockTime,
		&tx.Index,
		&tx.GasWanted,
		&tx.GasUsed,
		&tx.Fee,
		&tx.Memo,
		&tx.ErrorLog,
		&tx.Status,
		&tx.Codespace,
		&messagesJSON,
		&signersJSON,
		&tx.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction by hash %s: %w", hash, err)
	}

	if err := json.Unmarshal(messagesJSON, &tx.Messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction messages: %w", err)
	}

	if err := json.Unmarshal(signersJSON, &tx.Signers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction signers: %w", err)
	}

	return &tx, nil
}

func (r *PostgresRepository) GetTransactionsByBlockHeight(ctx context.Context, height int64) ([]*models.Transaction, error) {
	query := `
		SELECT 
			hash, block_height, block_time, index, gas_wanted, gas_used,
			fee, memo, error_log, status, codespace, messages, signers, created_at
		FROM transactions.transactions
		WHERE block_height = $1
		ORDER BY index ASC
	`

	rows, err := r.db.Query(ctx, query, height)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions for block %d: %w", height, err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var messagesJSON, signersJSON []byte

		err := rows.Scan(
			&tx.Hash,
			&tx.BlockHeight,
			&tx.BlockTime,
			&tx.Index,
			&tx.GasWanted,
			&tx.GasUsed,
			&tx.Fee,
			&tx.Memo,
			&tx.ErrorLog,
			&tx.Status,
			&tx.Codespace,
			&messagesJSON,
			&signersJSON,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		if err := json.Unmarshal(messagesJSON, &tx.Messages); err != nil {
			logger.Warn("Failed to unmarshal transaction messages",
				zap.String("hash", tx.Hash),
				zap.Error(err),
			)
		}

		if err := json.Unmarshal(signersJSON, &tx.Signers); err != nil {
			logger.Warn("Failed to unmarshal transaction signers",
				zap.String("hash", tx.Hash),
				zap.Error(err),
			)
		}

		transactions = append(transactions, &tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, nil
}

func (r *PostgresRepository) GetTransactionsByAddress(ctx context.Context, address string, limit, offset int) ([]*models.Transaction, error) {
	query := `
        SELECT 
            hash, block_height, block_time, index, gas_wanted, gas_used,
            fee, memo, error_log, status, codespace, messages, signers, created_at
        FROM transactions.transactions
        WHERE signers @> $1
        ORDER BY block_height DESC, index ASC
        LIMIT $2 OFFSET $3
    `

	// Create a JSON array with the address
	addressJSON, err := json.Marshal([]string{address})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal address for query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, addressJSON, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions for address %s: %w", address, err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var messagesJSON, signersJSON []byte

		err := rows.Scan(
			&tx.Hash,
			&tx.BlockHeight,
			&tx.BlockTime,
			&tx.Index,
			&tx.GasWanted,
			&tx.GasUsed,
			&tx.Fee,
			&tx.Memo,
			&tx.ErrorLog,
			&tx.Status,
			&tx.Codespace,
			&messagesJSON,
			&signersJSON,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		if err := json.Unmarshal(messagesJSON, &tx.Messages); err != nil {
			logger.Warn("Failed to unmarshal transaction messages",
				zap.String("hash", tx.Hash),
				zap.Error(err),
			)
		}

		if err := json.Unmarshal(signersJSON, &tx.Signers); err != nil {
			logger.Warn("Failed to unmarshal transaction signers",
				zap.String("hash", tx.Hash),
				zap.Error(err),
			)
		}

		transactions = append(transactions, &tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, nil
}

// Event operations
func (r *PostgresRepository) SaveEvent(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events.events (
			id, block_height, block_time, tx_hash, tx_index, event_type,
			event_index, module, attributes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (id) DO UPDATE SET
			block_height = EXCLUDED.block_height,
			block_time = EXCLUDED.block_time,
			tx_hash = EXCLUDED.tx_hash,
			tx_index = EXCLUDED.tx_index,
			event_type = EXCLUDED.event_type,
			event_index = EXCLUDED.event_index,
			module = EXCLUDED.module,
			attributes = EXCLUDED.attributes
	`

	attributesJSON, err := json.Marshal(event.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal event attributes: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		event.ID,
		event.BlockHeight,
		event.BlockTime,
		event.TxHash,
		event.TxIndex,
		event.EventType,
		event.EventIndex,
		event.Module,
		attributesJSON,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save event %s: %w", event.ID, err)
	}

	return nil
}

func (r *PostgresRepository) GetEventsByBlockHeight(ctx context.Context, height int64) ([]*models.Event, error) {
	query := `
        SELECT 
            id, block_height, block_time, tx_hash, tx_index, event_type,
            event_index, module, attributes, created_at
        FROM events.events
        WHERE block_height = $1
        ORDER BY tx_index ASC, event_index ASC
    `

	rows, err := r.db.Query(ctx, query, height)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for block %d: %w", height, err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var attributesJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.BlockHeight,
			&event.BlockTime,
			&event.TxHash,
			&event.TxIndex,
			&event.EventType,
			&event.EventIndex,
			&event.Module,
			&attributesJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		if err := json.Unmarshal(attributesJSON, &event.Attributes); err != nil {
			logger.Warn("Failed to unmarshal event attributes",
				zap.String("id", event.ID),
				zap.Error(err),
			)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

func (r *PostgresRepository) GetEventsByTxHash(ctx context.Context, txHash string) ([]*models.Event, error) {
	query := `
		SELECT 
			id, block_height, block_time, tx_hash, tx_index, event_type,
			event_index, module, attributes, created_at
		FROM events.events
		WHERE tx_hash = $1
		ORDER BY event_index ASC
	`

	rows, err := r.db.Query(ctx, query, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for transaction %s: %w", txHash, err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var attributesJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.BlockHeight,
			&event.BlockTime,
			&event.TxHash,
			&event.TxIndex,
			&event.EventType,
			&event.EventIndex,
			&event.Module,
			&attributesJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		if err := json.Unmarshal(attributesJSON, &event.Attributes); err != nil {
			logger.Warn("Failed to unmarshal event attributes",
				zap.String("id", event.ID),
				zap.Error(err),
			)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

func (r *PostgresRepository) GetEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*models.Event, error) {
	query := `
        SELECT 
            id, block_height, block_time, tx_hash, tx_index, event_type,
            event_index, module, attributes, created_at
        FROM events.events
        WHERE event_type = $1
        ORDER BY block_height DESC, tx_index ASC, event_index ASC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.Query(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by type %s: %w", eventType, err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var attributesJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.BlockHeight,
			&event.BlockTime,
			&event.TxHash,
			&event.TxIndex,
			&event.EventType,
			&event.EventIndex,
			&event.Module,
			&attributesJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		if err := json.Unmarshal(attributesJSON, &event.Attributes); err != nil {
			logger.Warn("Failed to unmarshal event attributes",
				zap.String("id", event.ID),
				zap.Error(err),
			)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

func (r *PostgresRepository) GetEventsByModule(ctx context.Context, module string, limit, offset int) ([]*models.Event, error) {
	query := `
        SELECT 
            id, block_height, block_time, tx_hash, tx_index, event_type,
            event_index, module, attributes, created_at
        FROM events.events
        WHERE module = $1
        ORDER BY block_height DESC, tx_index ASC, event_index ASC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.Query(ctx, query, module, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by module %s: %w", module, err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var attributesJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.BlockHeight,
			&event.BlockTime,
			&event.TxHash,
			&event.TxIndex,
			&event.EventType,
			&event.EventIndex,
			&event.Module,
			&attributesJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		if err := json.Unmarshal(attributesJSON, &event.Attributes); err != nil {
			logger.Warn("Failed to unmarshal event attributes",
				zap.String("id", event.ID),
				zap.Error(err),
			)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

// Account operations
func (r *PostgresRepository) SaveAccount(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts.accounts (
			address, account_type, balance, last_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		ON CONFLICT (address) DO UPDATE SET
			account_type = EXCLUDED.account_type,
			balance = EXCLUDED.balance,
			last_active = EXCLUDED.last_active,
			updated_at = EXCLUDED.updated_at
	`

	balanceJSON, err := json.Marshal(account.Balance)
	if err != nil {
		return fmt.Errorf("failed to marshal account balance: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		account.Address,
		account.AccountType,
		balanceJSON,
		account.LastActive,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save account %s: %w", account.Address, err)
	}

	return nil
}

func (r *PostgresRepository) GetAccountByAddress(ctx context.Context, address string) (*models.Account, error) {
	query := `
        SELECT 
            address, account_type, balance, last_active, created_at, updated_at
        FROM accounts.accounts
        WHERE address = $1
    `

	var account models.Account
	var balanceJSON []byte

	err := r.db.QueryRow(ctx, query, address).Scan(
		&account.Address,
		&account.AccountType,
		&balanceJSON,
		&account.LastActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get account by address %s: %w", address, err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(balanceJSON, &account.Balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account balance: %w", err)
	}

	return &account, nil
}

func (r *PostgresRepository) UpdateAccountBalance(ctx context.Context, address string, balance models.Coins) error {
	query := `
        UPDATE accounts.accounts
        SET balance = $1, 
            last_active = $2,
            updated_at = $3
        WHERE address = $4
    `

	// Convert balance to JSON
	balanceJSON, err := json.Marshal(balance)
	if err != nil {
		return fmt.Errorf("failed to marshal account balance: %w", err)
	}

	result, err := r.db.Exec(ctx, query,
		balanceJSON,
		time.Now().UTC(),
		time.Now().UTC(),
		address,
	)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	// Check if the account exists
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Dataset operations
func (r *PostgresRepository) SaveDataset(ctx context.Context, dataset *models.Dataset) error {
	query := `
		INSERT INTO datasets.datasets (
			denom, name, ticker, description, url, supply, categories,
			license_name, license_symbol, license_url, is_community_dataset,
			metadata, owner, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (denom) DO UPDATE SET
			name = EXCLUDED.name,
			ticker = EXCLUDED.ticker,
			description = EXCLUDED.description,
			url = EXCLUDED.url,
			supply = EXCLUDED.supply,
			categories = EXCLUDED.categories,
			license_name = EXCLUDED.license_name,
			license_symbol = EXCLUDED.license_symbol,
			license_url = EXCLUDED.license_url,
			is_community_dataset = EXCLUDED.is_community_dataset,
			metadata = EXCLUDED.metadata,
			owner = EXCLUDED.owner,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		dataset.Denom,
		dataset.Name,
		dataset.Ticker,
		dataset.Description,
		dataset.URL,
		dataset.Supply,
		dataset.Categories,
		dataset.LicenseName,
		dataset.LicenseSymbol,
		dataset.LicenseURL,
		dataset.IsCommunityDataset,
		dataset.Metadata,
		dataset.Owner,
		dataset.CreatedAt,
		dataset.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save dataset %s: %w", dataset.Denom, err)
	}

	return nil
}

func (r *PostgresRepository) GetDatasetByDenom(ctx context.Context, denom string) (*models.Dataset, error) {
	query := `
        SELECT 
            denom, name, ticker, description, url, supply, categories,
            license_name, license_symbol, license_url, is_community_dataset,
            metadata, owner, created_at, updated_at
        FROM datasets.datasets
        WHERE denom = $1
    `

	var dataset models.Dataset
	err := r.db.QueryRow(ctx, query, denom).Scan(
		&dataset.Denom,
		&dataset.Name,
		&dataset.Ticker,
		&dataset.Description,
		&dataset.URL,
		&dataset.Supply,
		&dataset.Categories,
		&dataset.LicenseName,
		&dataset.LicenseSymbol,
		&dataset.LicenseURL,
		&dataset.IsCommunityDataset,
		&dataset.Metadata,
		&dataset.Owner,
		&dataset.CreatedAt,
		&dataset.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get dataset by denom %s: %w", denom, err)
	}

	return &dataset, nil
}

func (r *PostgresRepository) GetDatasets(ctx context.Context, limit, offset int) ([]*models.Dataset, error) {
	query := `
        SELECT 
            denom, name, ticker, description, url, supply, categories,
            license_name, license_symbol, license_url, is_community_dataset,
            metadata, owner, created_at, updated_at
        FROM datasets.datasets
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query datasets: %w", err)
	}
	defer rows.Close()

	var datasets []*models.Dataset
	for rows.Next() {
		var dataset models.Dataset
		err := rows.Scan(
			&dataset.Denom,
			&dataset.Name,
			&dataset.Ticker,
			&dataset.Description,
			&dataset.URL,
			&dataset.Supply,
			&dataset.Categories,
			&dataset.LicenseName,
			&dataset.LicenseSymbol,
			&dataset.LicenseURL,
			&dataset.IsCommunityDataset,
			&dataset.Metadata,
			&dataset.Owner,
			&dataset.CreatedAt,
			&dataset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dataset row: %w", err)
		}
		datasets = append(datasets, &dataset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dataset rows: %w", err)
	}

	return datasets, nil
}

func (r *PostgresRepository) UpdateDataset(ctx context.Context, dataset *models.Dataset) error {
	// TODO
	return nil
}

func (r *PostgresRepository) SaveDatasetContribution(ctx context.Context, contribution *models.DatasetContribution) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetDatasetContributions(ctx context.Context, denom string) ([]*models.DatasetContribution, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) SavePendingDatasetContribution(ctx context.Context, contribution *models.PendingDatasetContribution) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetPendingDatasetContributions(ctx context.Context, denom string) ([]*models.PendingDatasetContribution, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) DeletePendingDatasetContribution(ctx context.Context, contributionID string) error {
	// TODO
	return nil
}

// Validator operations
func (r *PostgresRepository) SaveValidator(ctx context.Context, validator *models.Validator) error {
	query := `
		INSERT INTO validators.validators (
			consensus_address, operator_address, account_address, moniker, identity,
			website, security_contact, details, commission_rate, voting_power,
			status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (consensus_address) DO UPDATE SET
			operator_address = EXCLUDED.operator_address,
			account_address = EXCLUDED.account_address,
			moniker = EXCLUDED.moniker,
			identity = EXCLUDED.identity,
			website = EXCLUDED.website,
			security_contact = EXCLUDED.security_contact,
			details = EXCLUDED.details,
			commission_rate = EXCLUDED.commission_rate,
			voting_power = EXCLUDED.voting_power,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		validator.ConsensusAddress,
		validator.OperatorAddress,
		validator.AccountAddress,
		validator.Moniker,
		validator.Identity,
		validator.Website,
		validator.SecurityContact,
		validator.Details,
		validator.CommissionRate,
		validator.VotingPower,
		validator.Status,
		validator.CreatedAt,
		validator.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save validator %s: %w", validator.ConsensusAddress, err)
	}

	return nil
}

func (r *PostgresRepository) GetValidatorByConsensusAddress(ctx context.Context, consensusAddress string) (*models.Validator, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetValidatorByOperatorAddress(ctx context.Context, operatorAddress string) (*models.Validator, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetValidators(ctx context.Context, limit, offset int) ([]*models.Validator, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateValidator(ctx context.Context, validator *models.Validator) error {
	// TODO
	return nil
}

func (r *PostgresRepository) SaveValidatorUptime(ctx context.Context, uptime *models.ValidatorUptime) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetValidatorUptime(ctx context.Context, consensusAddress string, startHeight, endHeight int64) ([]*models.ValidatorUptime, error) {
	// TODO
	return nil, nil
}

// Validator rewards
func (r *PostgresRepository) SaveValidatorReward(ctx context.Context, reward *models.ValidatorReward) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetValidatorRewards(ctx context.Context, validatorAddress string, startHeight, endHeight int64) ([]*models.ValidatorReward, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetDelegatorRewards(ctx context.Context, delegatorAddress string, startHeight, endHeight int64) ([]*models.ValidatorReward, error) {
	// TODO
	return nil, nil
}

// Validator slashing
func (r *PostgresRepository) SaveSlashingEvent(ctx context.Context, slashing *models.SlashingEvent) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetSlashingEvents(ctx context.Context, validatorAddress string, startHeight, endHeight int64) ([]*models.SlashingEvent, error) {
	// TODO
	return nil, nil
}

// Delegations
func (r *PostgresRepository) SaveDelegation(ctx context.Context, delegation *models.Delegation) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetDelegationsByDelegator(ctx context.Context, delegatorAddress string) ([]*models.Delegation, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetDelegationsByValidator(ctx context.Context, validatorAddress string) ([]*models.Delegation, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateDelegation(ctx context.Context, delegation *models.Delegation) error {
	// TODO
	return nil
}

// Governance operations
func (r *PostgresRepository) SaveProposal(ctx context.Context, proposal *models.Proposal) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetProposalByID(ctx context.Context, proposalID int64) (*models.Proposal, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetProposals(ctx context.Context, limit, offset int) ([]*models.Proposal, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateProposal(ctx context.Context, proposal *models.Proposal) error {
	// TODO
	return nil
}

func (r *PostgresRepository) SaveProposalDeposit(ctx context.Context, deposit *models.ProposalDeposit) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetProposalDeposits(tx context.Context, proposalID int64) ([]*models.ProposalDeposit, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) SaveProposalVote(ctx context.Context, vote *models.ProposalVote) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetProposalVotes(ctx context.Context, proposalID int64, limit, offset int) ([]*models.ProposalVote, error) {
	// TODO
	return nil, nil
}

// WASM operations
func (r *PostgresRepository) SaveWasmCode(ctx context.Context, code *models.WasmCode) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetWasmCodeByID(ctx context.Context, codeID int64) (*models.WasmCode, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetWasmCodes(ctx context.Context, limit, offset int) ([]*models.WasmCode, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) SaveWasmContract(ctx context.Context, contract *models.WasmContract) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetWasmContractByAddress(ctx context.Context, address string) (*models.WasmContract, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetWasmContractsByCodeID(ctx context.Context, codeID int64) ([]*models.WasmContract, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateWasmContract(ctx context.Context, contract *models.WasmContract) error {
	// TODO
	return nil
}

func (r *PostgresRepository) SaveWasmExecute(ctx context.Context, execute *models.WasmExecute) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetWasmExecutesByContract(ctx context.Context, contractAddress string, limit, offset int) ([]*models.WasmExecute, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) SaveWasmQuery(ctx context.Context, query *models.WasmQuery) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetWasmQueriesByContract(ctx context.Context, contractAddress string, limit, offset int) ([]*models.WasmQuery, error) {
	// TODO
	return nil, nil
}

// IBC operations
func (r *PostgresRepository) SaveIBCChannel(ctx context.Context, channel *models.IBCChannel) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetIBCChannelByID(ctx context.Context, channelID, portID string) (*models.IBCChannel, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetIBCChannels(ctx context.Context, limit, offset int) ([]*models.IBCChannel, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateIBCChannel(ctx context.Context, channel *models.IBCChannel) error {
	// TODO
	return nil
}

func (r *PostgresRepository) SaveIBCConnection(ctx context.Context, connection *models.IBCConnection) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetIBCConnectionByID(ctx context.Context, connectionID string) (*models.IBCConnection, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetIBCConnections(ctx context.Context, limit, offset int) ([]*models.IBCConnection, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateIBCConnection(ctx context.Context, connection *models.IBCConnection) error {
	// TODO
	return nil
}

func (r *PostgresRepository) SaveIBCClient(ctx context.Context, client *models.IBCClient) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetIBCClientByID(ctx context.Context, clientID string) (*models.IBCClient, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetIBCClients(ctx context.Context, limit, offset int) ([]*models.IBCClient, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateIBCClient(ctx context.Context, client *models.IBCClient) error {
	// TODO
	return nil
}

func (r *PostgresRepository) SaveIBCPacket(ctx context.Context, packet *models.IBCPacket) error {
	// TODO
	return nil
}

func (r *PostgresRepository) GetIBCPacketByID(ctx context.Context, id string) (*models.IBCPacket, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) GetIBCPacketsByChannel(ctx context.Context, channelID, portID string, limit, offset int) ([]*models.IBCPacket, error) {
	// TODO
	return nil, nil
}

func (r *PostgresRepository) UpdateIBCPacket(ctx context.Context, packet *models.IBCPacket) error {
	// TODO
	return nil
}
