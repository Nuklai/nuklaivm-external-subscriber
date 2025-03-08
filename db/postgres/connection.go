// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"

	"github.com/nuklai/nuklaivm-external-subscriber/config"
)

type Database struct {
	pool *pgxpool.Pool
	cfg  *config.DatabaseConfig
}

// Create a new psql connection
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	// Connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set pool config
	poolConfig.MaxConns = int32(cfg.MaxOpenConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConnections)
	poolConfig.MaxConnLifetime = cfg.ConnectionMaxLifetime
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Png db
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to postgres database",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
		zap.Int("max_connections", cfg.MaxOpenConnections),
	)

	return &Database{
		pool: pool,
		cfg:  cfg,
	}, nil
}

// Close the db connection
func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
		logger.Info("Database connection closed")
	}
}

// Returns the connection pool
func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

// Check db connection
func (db *Database) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Run the db migrations
// TODO: use a migration lib: golang-migrate/migrate
func (db *Database) RunMigrations(migrationsPath string) error {
	logger.Info("Running database migrations", zap.String("path", migrationsPath))
	// TODO: handle db migration
	return nil
}
