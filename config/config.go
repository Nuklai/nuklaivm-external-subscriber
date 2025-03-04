// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Indexer initial configs
type Config struct {
	LogLevel   string           `mapstructure:"log_level"`
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
}

// HTTP Server initial configs
type ServerConfig struct {
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	EnableTLS          bool   `mapstructure:"enable_tls"`
	CertFile           string `mapstructure:"cert_file"`
	KeyFile            string `mapstructure:"key_file"`
	EnableRateLimiting bool   `mapstructure:"enable_rate_limiting"`
	RequestsPerMinute  int    `mapstructure:"requests_per_minute"`
}

// Database initial configs
type DatabaseConfig struct {
	Host                  string        `mapstructure:"host"`
	Port                  int           `mapstructure:"port"`
	Username              string        `mapstructure:"username"`
	Password              string        `mapstructure:"password"`
	Database              string        `mapstructure:"database"`
	SSLMode               string        `mapstructure:"ssl_mode"`
	MaxOpenConnections    int           `mapstructure:"max_open_connections"`
	MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `mapstructure:"connection_max_lifetime"`
}

// Chain initial configs
type BlockchainConfig struct {
	RPCEndpoints    []string      `mapstructure:"rpc_endpoints"`
	StartHeight     int64         `mapstructure:"start_height"`
	SyncConcurrency int           `mapstructure:"sync_concurrency"`
	FetchInterval   time.Duration `mapstructure:"fetch_interval"`
	FetchTimeout    time.Duration `mapstructure:"fetch_timeout"`
}

// Loads the config from the specified yaml file.
func Load(configPath string) (*Config, error) {
	configFile := filepath.Base(configPath)
	configDir := filepath.Dir(configPath)

	v := viper.New()
	v.SetConfigName(configFile[:len(configFile)-len(filepath.Ext(configFile))])
	v.AddConfigPath(configDir)
	v.SetConfigType("yaml")

	// Set defaults
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read the config file: %w", err)
		}
	}

	// Override here
	v.SetEnvPrefix("NUKLAI_INDEXER")
	v.AutomaticEnv()

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// Default config values
func setDefaults(v *viper.Viper) {
	v.SetDefault("log_level", "info")

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.enable_tls", false)
	v.SetDefault("server.cert_file", "")
	v.SetDefault("server.key_file", "")
	v.SetDefault("server.enable_rate_limiting", true)
	v.SetDefault("server.requests_per_minute", 100)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.database", "nuklai_indexer")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_connections", 25)
	v.SetDefault("database.max_idle_connections", 5)
	v.SetDefault("database.connection_max_lifetime", "1h")

	// Blockchain defaults
	v.SetDefault("blockchain.rpc_endpoints", []string{"http://localhost:26657"})
	v.SetDefault("blockchain.start_height", 1)
	v.SetDefault("blockchain.sync_concurrency", 10)
	v.SetDefault("blockchain.fetch_interval", "1s")
	v.SetDefault("blockchain.fetch_timeout", "10s")
}

// Returns the db connection string
func (dc *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dc.Host, dc.Port, dc.Username, dc.Password, dc.Database, dc.SSLMode,
	)
}

// Saves the default config to specified path
func SaveDefault(configPath string) error {
	// New viper instance
	v := viper.New()

	// Set defaults
	setDefaults(v)

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
