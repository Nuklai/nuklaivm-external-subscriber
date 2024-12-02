package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// GetAllTransactions retrieves all transactions with pagination
// GetAllTransactions retrieves all transactions with pagination and total count
func GetAllTransactions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of transactions
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count transactions"})
			return
		}

		// Fetch paginated transactions
		transactions, err := models.FetchAllTransactions(db, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions"})
			return
		}

		// Return response with counter
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions"})
			return
		}

		c.JSON(http.StatusOK, transactions)
	}
}
