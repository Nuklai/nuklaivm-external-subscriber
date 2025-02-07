// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
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
	IsReachable         bool      `json:"is_reachable"`
	LastError           string    `json:"last_error,omitempty"`
	LastChecked         time.Time `json:"last_checked"`
	LastSuccessful      time.Time `json:"last_successful"`
	ResponseTime        string    `json:"response_time"`
	ResponseTimeSeconds float64   `json:"response_time_seconds"`
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
	ServiceStatuse  map[string]*ServiceStatus `json:"service_statuse"`
	BlockchainStats *BlockchainStats          `json:"blockchain_stats"`
	CurrentIncident *HealthEvent              `json:"current_incident"`
}

type DailyHealthSummary struct {
	Date      time.Time   `json:"date"`
	State     HealthState `json:"state"`
	Incidents []string    `json:"incidents"`
}

func UpdateDailyHealthSummary(db *sql.DB, currentStatus HealthStatus) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			if err != nil {
				tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}
	}()

	today := time.Now().UTC().Truncate(24 * time.Hour)

	var currentState HealthState
	var incidents []string

	if currentStatus.State == HealthStateRed {
		currentState = HealthStateRed
	} else if currentStatus.State == HealthStateYellow {
		currentState = HealthStateYellow
	} else {
		currentState = HealthStateGreen
	}

	if currentStatus.CurrentIncident != nil {
		incidents = []string{currentStatus.CurrentIncident.Description}
	}

	_, execErr := tx.Exec(`
        INSERT INTO daily_health_summaries (date, state, incidents, last_updated)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (date) DO UPDATE
        SET state = CASE 
            WHEN daily_health_summaries.state = 'red' OR EXCLUDED.state = 'red' THEN 'red'
            WHEN daily_health_summaries.state = 'yellow' OR EXCLUDED.state = 'yellow' THEN 'yellow'
            ELSE 'green'
        END,
        incidents = CASE
            WHEN EXCLUDED.incidents IS NOT NULL AND EXCLUDED.incidents != '{}'::text[]
            THEN EXCLUDED.incidents  -- Keep only latest incident
            ELSE daily_health_summaries.incidents
        END,
        last_updated = NOW()`,
		today, currentState, pq.Array(incidents))

	if execErr != nil {
		err = execErr
		return err
	}

	return nil
}

func Fetch90DayHealth(db *sql.DB) ([]DailyHealthSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
        SELECT date, state, incidents 
        FROM daily_health_summaries
        WHERE date > NOW() - INTERVAL '90 days'
        ORDER BY date DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query health summaries: %w", err)
	}
	defer rows.Close()

	var summaries []DailyHealthSummary
	for rows.Next() {
		var summary DailyHealthSummary
		if err := rows.Scan(&summary.Date, &summary.State,
			pq.Array(&summary.Incidents)); err != nil {
			return nil, fmt.Errorf("failed to scan health summary: %w", err)
		}
		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	return summaries, nil
}
