package api

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

type HealthMonitor struct {
	db            *sql.DB
	mu            sync.RWMutex
	currentStatus models.HealthStatus
	grpcPort      string
}

// InitHealthMonitor initializes the health monitor
func InitHealthMonitor(db *sql.DB, grpcPort string) *HealthMonitor {
	monitor := &HealthMonitor{
		db:       db,
		grpcPort: grpcPort,
		currentStatus: models.HealthStatus{
			State:           models.HealthStateGreen,
			Details:         make(map[string]bool),
			ServiceStatuses: make(map[string]*models.ServiceStatus),
			BlockchainStats: &models.BlockchainStats{},
			CurrentIncident: nil,
		},
	}
	return monitor
}

// FetchBlockchainHealth retrieves the current VM health stats
func (h *HealthMonitor) FetchBlockchainHealth() (*models.ServiceStatus, *models.BlockchainStats) {
	start := time.Now()
	status := &models.ServiceStatus{
		LastChecked: time.Now().UTC(),
	}
	stats := &models.BlockchainStats{}

	// Get the latest block info
	var lastBlock struct {
		Height    int64
		Hash      string
		Timestamp time.Time
	}

	err := h.db.QueryRow(`
        SELECT block_height, block_hash, timestamp 
        FROM blocks 
        ORDER BY block_height DESC 
        LIMIT 1
    `).Scan(&lastBlock.Height, &lastBlock.Hash, &lastBlock.Timestamp)

	if err != nil {
		status.IsReachable = false
		status.LastError = fmt.Sprintf("Failed to query last block: %v", err)
		return status, stats
	}

	// 6/12 secinds block tracking
	blockAge := time.Since(lastBlock.Timestamp)
	if blockAge > 12*time.Second {
		status.IsReachable = false
		status.LastError = fmt.Sprintf("Connection to NuklaiVM  lost - no new blocks in %v",
			blockAge.Round(time.Second))
	} else {
		status.IsReachable = true
		status.LastSuccessful = time.Now().UTC()
	}

	var blockCount, txCount int
	err = h.db.QueryRow(`
        SELECT 
            COUNT(DISTINCT blocks.block_height),
            COUNT(DISTINCT transactions.tx_hash)
        FROM blocks 
        LEFT JOIN transactions ON blocks.block_hash = transactions.block_hash
        WHERE blocks.timestamp > NOW() - INTERVAL '1 minute'
    `).Scan(&blockCount, &txCount)

	if err != nil {
		log.Printf("Error counting recent blocks/transactions: %v", err)
	}

	// Calc average block time from the last minute
	var avgBlockTime float64
	err = h.db.QueryRow(`
        WITH block_times AS (
            SELECT 
                block_height,
                timestamp,
                LAG(timestamp) OVER (ORDER BY block_height DESC) as prev_timestamp
            FROM blocks
            WHERE timestamp > NOW() - INTERVAL '1 minute'
        )
        SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (timestamp - prev_timestamp))), 0)
        FROM block_times
        WHERE prev_timestamp IS NOT NULL
    `).Scan(&avgBlockTime)

	if err != nil {
		log.Printf("Error calculating average block time: %v", err)
		avgBlockTime = 0
	}

	stats.LastBlockHeight = lastBlock.Height
	stats.LastBlockHash = lastBlock.Hash
	stats.LastBlockTime = lastBlock.Timestamp
	stats.ConsensusActive = blockAge <= 12*time.Second

	status.ResponseTime = time.Since(start).String()
	return status, stats
}

// UpdateHealthState updates the health state and tracks incidents
func (h *HealthMonitor) UpdateHealthState(newState models.HealthState, description string, serviceNames []string) {
	if newState == models.HealthStateGreen {
		// If green close any current incident
		if h.currentStatus.CurrentIncident != nil {
			now := time.Now().UTC()
			_, err := h.db.Exec(`
                UPDATE health_events 
                SET end_time = $1, 
                    duration = EXTRACT(EPOCH FROM ($1 - start_time))::INT
                WHERE id = $2 AND end_time IS NULL`,
				now, h.currentStatus.CurrentIncident.ID)
			if err != nil {
				log.Printf("Error updating health event: %v", err)
			}

			h.currentStatus.CurrentIncident = nil
		}
		h.currentStatus.State = newState
		return
	}

	if newState != models.HealthStateGreen {
		var lastIncident models.HealthEvent
		err := h.db.QueryRow(`
        SELECT id, state, description, service_names, start_time, timestamp 
        FROM health_events 
        WHERE state = $1 AND end_time IS NULL
        ORDER BY start_time DESC LIMIT 1`,
			newState).Scan(
			&lastIncident.ID,
			&lastIncident.State,
			&lastIncident.Description,
			pq.Array(&lastIncident.ServiceNames),
			&lastIncident.StartTime,
			&lastIncident.Timestamp)

		shouldCreateNew := true
		if err == nil {
			timeSinceLastIncident := time.Since(lastIncident.StartTime)
			if timeSinceLastIncident < time.Hour {
				shouldCreateNew = false
				h.currentStatus.CurrentIncident = &lastIncident
			}
		}

		if shouldCreateNew {
			event := &models.HealthEvent{
				State:        newState,
				Description:  description,
				ServiceNames: serviceNames,
				StartTime:    time.Now().UTC(),
				Timestamp:    time.Now().UTC(),
			}

			err := h.db.QueryRow(`
                INSERT INTO health_events (
                    state, description, service_names, start_time, timestamp
                ) VALUES ($1, $2, $3, $4, $5)
                RETURNING id`,
				event.State, event.Description, pq.Array(serviceNames),
				event.StartTime, event.Timestamp).Scan(&event.ID)

			if err != nil {
				log.Printf("Error creating health event: %v", err)
			}
			h.currentStatus.CurrentIncident = event
		}
	}

	h.currentStatus.State = newState
}

// GetHealthStatus returns the complete health status of the VM
func (h *HealthMonitor) GetHealthStatus() models.HealthStatus {
	h.mu.Lock()
	defer h.mu.Unlock()

	blockchainStatus, blockchainStats := h.FetchBlockchainHealth()

	h.currentStatus.Details = map[string]bool{
		"blockchain": blockchainStatus.IsReachable,
	}

	h.currentStatus.ServiceStatuses = map[string]*models.ServiceStatus{
		"blockchain": blockchainStatus,
	}

	h.currentStatus.BlockchainStats = blockchainStats

	if !blockchainStatus.IsReachable {
		description := strings.Builder{}
		description.WriteString("CRITICAL: NuklaiVM Unresponsive\n")
		description.WriteString(fmt.Sprintf("- Error: %s\n", blockchainStatus.LastError))
		description.WriteString(fmt.Sprintf("- Last Block Height: %d\n", blockchainStats.LastBlockHeight))
		description.WriteString(fmt.Sprintf("- Last Block Time: %s\n", blockchainStats.LastBlockTime.Format(time.RFC3339)))
		description.WriteString(fmt.Sprintf("- Block Age: %v\n", time.Since(blockchainStats.LastBlockTime).Round(time.Second)))

		h.UpdateHealthState(models.HealthStateRed, description.String(), []string{"blockchain"})
	} else if h.currentStatus.State != models.HealthStateGreen {
		h.UpdateHealthState(models.HealthStateGreen, "", nil)
	}

	return h.currentStatus
}
