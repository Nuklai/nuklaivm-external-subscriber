// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nuklai/nuklaivm-external-subscriber/core"

	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	defaultConfigPath = "./config/config.yaml"
)

var (
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "indexer",
		Short: "Nuklai blockchain indexer",
		Long:  `Nuklai blockchain indexer captures, stores, and exposes blockchain data for querying and analysis.`,
		Run:   run,
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", defaultConfigPath, "path to config file")

	// Init: Generate default config
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the indexer with default configuration",
		Long:  `Generate a default configuration file for the indexer.`,
		Run:   initConfig,
	}
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig(_ *cobra.Command, _ []string) {
	// Create config dir if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveDefault(configPath); err != nil {
		fmt.Printf("Error saving default config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Default configuration saved to %s\n", configPath)
}

func run(_ *cobra.Command, _ []string) {
	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Init logger
	logger.InitLogger(cfg.LogLevel)
	defer logger.Logger.Sync()

	logger.Info("Starting Nuklai indexer...",
		zap.String("config_path", configPath),
		zap.String("log_level", cfg.LogLevel),
	)

	logger.Info("Configuration loaded",
		zap.String("server_host", cfg.Server.Host),
		zap.Int("server_port", cfg.Server.Port),
		zap.Strings("rpc_endpoints", cfg.Blockchain.RPCEndpoints),
	)

	// Initialize indexer core
	indexer, err := core.NewIndexer(cfg)
	if err != nil {
		logger.Fatal("Failed to create indexer", zap.Error(err))
	}

	// Create context with timeout, use the cancel function
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start indexer
	if err := indexer.Start(ctx); err != nil {
		logger.Fatal("Failed to start indexer", zap.Error(err))
	}

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Nuklai indexer...")

	// Create a new context with timeout for shutdown
	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// TODO: Use shutdownCtx

	indexer.Stop()
}
