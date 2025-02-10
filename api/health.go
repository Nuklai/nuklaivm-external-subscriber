// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/lib/pq"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// GetHealth retrieves the current health status of the system
func GetHealth(monitor *HealthMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := monitor.GetHealthStatus()

		response := gin.H{
			"api_healthy": err == nil,
			"status":      nil,
		}

		if err == nil {
			response["status"] = status
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetHealthHistory retrieves historical health events for insidents
func GetHealthHistory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
            SELECT id, state, description, service_names, 
                   start_time, end_time, COALESCE(duration, 0), timestamp 
            FROM health_events 
            ORDER BY timestamp DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch health history"})
			return
		}
		defer rows.Close()

		var events []models.HealthEvent
		for rows.Next() {
			var event models.HealthEvent
			var endTime sql.NullTime
			var serviceNamesArray pq.StringArray
			var duration sql.NullInt64

			err := rows.Scan(
				&event.ID,
				&event.State,
				&event.Description,
				&serviceNamesArray,
				&event.StartTime,
				&endTime,
				&duration,
				&event.Timestamp,
			)
			if err != nil {
				log.Printf("Error scanning health event: %v", err)
				continue
			}

			if endTime.Valid {
				event.EndTime = &endTime.Time
			}
			if duration.Valid {
				event.Duration = duration.Int64
			}
			event.ServiceNames = []string(serviceNamesArray)
			events = append(events, event)
		}

		c.JSON(http.StatusOK, events)
	}
}

// Get90DayHealth retrives a 90day health summary
func Get90DayHealth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		summaries, err := models.Fetch90DayHealth(db)
		if err != nil {
			log.Printf("Error fetching 90-day health history: %v", err)
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Unable to retrieve health history"})
			return
		}
		c.JSON(http.StatusOK, summaries)
	}
}
