// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"time"
)

type HealthState string

const (
	HealthStateGreen  HealthState = "green"
	HealthStateYellow HealthState = "yellow"
	HealthStateRed    HealthState = "red"
)

type BlockchainStats struct {
	LastBlockHeight int64     `json:"last_block_height"`
	LastBlockHash   string    `json:"last_block_hash"`
	LastBlockTime   time.Time `json:"last_block_time"`
	ConsensusActive bool      `json:"consensus_active"`
}

type ServiceStatus struct {
	IsReachable    bool      `json:"is_reachable"`
	LastError      string    `json:"last_error,omitempty"`
	LastChecked    time.Time `json:"last_checked"`
	LastSuccessful time.Time `json:"last_successful"`
	ResponseTime   string    `json:"response_time"`
}

type HealthEvent struct {
	ID           int64       `json:"id"`
	State        HealthState `json:"state"`
	Description  string      `json:"description"`
	ServiceNames []string    `json:"service_names"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      *time.Time  `json:"end_time"`
	Duration     int64       `json:"duration"`
	Timestamp    time.Time   `json:"timestamp"`
}

type HealthStatus struct {
	State           HealthState               `json:"state"`
	Details         map[string]bool           `json:"details"`
	ServiceStatuses map[string]*ServiceStatus `json:"service_statuses"`
	BlockchainStats *BlockchainStats          `json:"blockchain_stats"`
	CurrentIncident *HealthEvent              `json:"current_incident"`
}
