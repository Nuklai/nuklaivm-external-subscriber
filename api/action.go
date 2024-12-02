package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// GetAllActions retrieves all actions with pagination and total count
func GetAllActions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of actions
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM actions`).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count actions"})
			return
		}

		// Fetch paginated actions
		actions, err := models.FetchAllActions(db, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve actions"})
			return
		}

		// Return response with counter
		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   actions,
		})
	}
}

// GetActionsByBlock retrieves actions by block height or hash
func GetActionsByBlock(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		blockIdentifier := c.Param("identifier")

		actions, err := models.FetchActionsByBlock(db, blockIdentifier)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve actions"})
			return
		}

		c.JSON(http.StatusOK, actions)
	}
}

// GetActionsByTransactionHash retrieves actions by transaction hash
func GetActionsByTransactionHash(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		txHash := c.Param("tx_hash")

		actions, err := models.FetchActionsByTransactionHash(db, txHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve actions"})
			return
		}

		c.JSON(http.StatusOK, actions)
	}
}
