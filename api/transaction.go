// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// GetAllTransactions retrieves all transactions with pagination and supports additional filters
func GetAllTransactions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		txHash := c.Query("tx_hash")
		blockHash := c.Query("block_hash")
		actionType := c.Query("action_type")
		actionName := strings.ToLower(c.Query("action_name"))
		user := c.Query("user")
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Normalize user to search with and without "0x" prefix
		if user != "" {
			user = strings.TrimPrefix(user, "0x")
		}

		// Get total count with filters
		totalCount, err := models.CountFilteredTransactions(db, txHash, blockHash, actionType, actionName, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count transactions"})
			return
		}

		// Fetch filtered transactions with pagination
		transactions, err := models.FetchFilteredTransactions(db, txHash, blockHash, actionType, actionName, user, limit, offset)
		if err != nil {
			log.Printf("Error fetching transactions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   transactions,
		})
	}
}

// GetTransactionByHash retrieves a transaction by its hash
func GetTransactionByHash(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		txHash := c.Param("tx_hash")

		transaction, err := models.FetchTransactionByHash(db, txHash)
		if err != nil {
			log.Printf("Error fetching transaction: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}

		c.JSON(http.StatusOK, transaction)
	}
}

// GetTransactionsByBlock retrieves transactions associated with a block by height or hash
func GetTransactionsByBlock(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		blockIdentifier := c.Param("identifier")

		transactions, err := models.FetchTransactionsByBlock(db, blockIdentifier)
		if err != nil {
			log.Printf("Error fetching transactions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions"})
			return
		}

		c.JSON(http.StatusOK, transactions)
	}
}

// GetTransactionsByUser retrieves transactions for a specific user (sponsor or actor or receiver)
func GetTransactionsByUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.Param("user")
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of user's transactions
		var totalCount int
		err := db.QueryRow(`
            SELECT COUNT(*)
            FROM transactions
            WHERE
								sponsor ILIKE $1
								OR EXISTS (
									SELECT 1
									FROM unnest(actors) AS actor
									WHERE actor ILIKE $1
								)
								OR EXISTS (
									SELECT 1
									FROM unnest(receivers) AS receiver
									WHERE receiver ILIKE $1
								)
        `, "%"+user+"%").Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count transactions for user"})
			return
		}

		// Fetch paginated transactions for the user
		transactions, err := models.FetchTransactionsByUser(db, user, limit, offset)
		if err != nil {
			log.Printf("Error fetching transactions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions for user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   transactions,
		})
	}
}

// GetTransactionVolumes retrieves all actions volumes by 12hrs, 24hrs, 7days & 30 days
func GetAllActionVolumes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		volumes, err := models.FetchAllActionVolumes(db)
		if err != nil {
			log.Printf("Error fetching action volumes: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve action volumes"})
			return
		}

		c.JSON(http.StatusOK, volumes)
	}
}

// GetTotalTransferVolume retrieves the all-time total transfer value
func GetTotalTransferVolume(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		volume, err := models.FetchTotalTransferVolume(db)
		if err != nil {
			log.Printf("Error fetching total transfer value: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve total transfer value"})
			return
		}

		c.JSON(http.StatusOK, volume)
	}
}

// GetActionVolumesByName retrieves an actions volume by it's name
func GetActionVolumesByName(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		actionName := c.Param("action_name")

		volume, err := models.FetchActionVolumesByName(db, actionName)
		if err != nil {
			log.Printf("Error fetching action volumes: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve action volumes"})
			return
		}

		c.JSON(http.StatusOK, volume)
	}
}

// GetEstimatedFeeByActionType retrieves the estimated fee for a specific action type
func GetEstimatedFeeByActionType(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		actionType := c.Param("action_type")
		interval := c.DefaultQuery("interval", "1m")

		result, err := calculateEstimatedFee(db, "action_type", actionType, interval)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

// GetEstimatedFeeByActionName retrieves the estimated fee for a specific action name
func GetEstimatedFeeByActionName(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		actionName := c.Param("action_name")
		interval := c.DefaultQuery("interval", "1m")

		result, err := calculateEstimatedFee(db, "LOWER(action_name)", strings.ToLower(actionName), interval)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func GetAggregateEstimatedFees(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		interval := c.DefaultQuery("interval", "1m")

		// Ensure both database timestamps and query intervals are in UTC
		rows, err := db.Query(`
            SELECT
                a.action_type,
                a.action_name,
                COALESCE(AVG(t.fee), 0) AS avg_fee,
                COALESCE(MIN(t.fee), 0) AS min_fee,
                COALESCE(MAX(t.fee), 0) AS max_fee,
                COUNT(*) AS tx_count
            FROM actions a
            JOIN transactions t ON a.tx_hash = t.tx_hash
            WHERE t.timestamp AT TIME ZONE 'UTC' >= (NOW() AT TIME ZONE 'UTC') - $1::interval
            GROUP BY a.action_type, a.action_name`, interval)
		if err != nil {
			log.Printf("SQL Query Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve estimated fees"})
			return
		}
		defer rows.Close()

		var aggregated struct {
			AvgFee  float64 `json:"avg_fee"`
			MinFee  float64 `json:"min_fee"`
			MaxFee  float64 `json:"max_fee"`
			TxCount int     `json:"tx_count"`
			Fees    []gin.H `json:"fees"`
		}

		for rows.Next() {
			var item struct {
				ActionType int     `json:"action_type"`
				ActionName string  `json:"action_name"`
				AvgFee     float64 `json:"avg_fee"`
				MinFee     float64 `json:"min_fee"`
				MaxFee     float64 `json:"max_fee"`
				TxCount    int     `json:"tx_count"`
			}
			if err := rows.Scan(&item.ActionType, &item.ActionName, &item.AvgFee, &item.MinFee, &item.MaxFee, &item.TxCount); err != nil {
				log.Printf("Row Scan Error: %v", err)
				continue
			}

			aggregated.Fees = append(aggregated.Fees, gin.H{
				"action_type": item.ActionType,
				"action_name": item.ActionName,
				"avg_fee":     item.AvgFee,
				"min_fee":     item.MinFee,
				"max_fee":     item.MaxFee,
				"tx_count":    item.TxCount,
			})
			aggregated.AvgFee += item.AvgFee * float64(item.TxCount) // Weighted average
			aggregated.TxCount += item.TxCount
			if aggregated.MinFee == 0 || item.MinFee < aggregated.MinFee {
				aggregated.MinFee = item.MinFee
			}
			if item.MaxFee > aggregated.MaxFee {
				aggregated.MaxFee = item.MaxFee
			}
		}

		if aggregated.TxCount > 0 {
			aggregated.AvgFee /= float64(aggregated.TxCount) // Final weighted average
		}

		c.JSON(http.StatusOK, aggregated)
	}
}

func calculateEstimatedFee(db *sql.DB, column, value, interval string) (gin.H, error) {
	var result struct {
		AvgFee     sql.NullFloat64
		MinFee     sql.NullFloat64
		MaxFee     sql.NullFloat64
		TxCount    int
		ActionType int
		ActionName string
	}

	query := `
        SELECT
            COALESCE(AVG(t.fee), 0) AS avg_fee,
            COALESCE(MIN(t.fee), 0) AS min_fee,
            COALESCE(MAX(t.fee), 0) AS max_fee,
            COUNT(*) AS tx_count,
            a.action_type,
            a.action_name
        FROM actions a
        JOIN transactions t ON a.tx_hash = t.tx_hash
        WHERE ` + column + ` = $1 AND t.timestamp AT TIME ZONE 'UTC' >= (NOW() AT TIME ZONE 'UTC') - $2::interval
        GROUP BY a.action_type, a.action_name`

	err := db.QueryRow(query, value, interval).Scan(
		&result.AvgFee, &result.MinFee, &result.MaxFee,
		&result.TxCount, &result.ActionType, &result.ActionName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty values if no rows are found
			return gin.H{
				"avg_fee":  0.0,
				"min_fee":  0.0,
				"max_fee":  0.0,
				"tx_count": 0,
			}, nil
		}
		log.Printf("QueryRow Error: %v", err)
		return nil, err
	}

	return gin.H{
		"avg_fee":     result.AvgFee.Float64,
		"min_fee":     result.MinFee.Float64,
		"max_fee":     result.MaxFee.Float64,
		"tx_count":    result.TxCount,
		"action_type": result.ActionType,
		"action_name": result.ActionName,
	}, nil
}
