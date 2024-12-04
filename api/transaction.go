package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

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
			log.Printf("Error fetching transactions: %v", err)
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

// GetTransactionsByUser retrieves transactions for a specific user (sponsor)
func GetTransactionsByUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.Param("user")

		// Normalize the user identifier by removing "0x" prefix if present
		user = strings.TrimPrefix(user, "0x")

		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of user's transactions
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM transactions WHERE sponsor = $1`, user).Scan(&totalCount)
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
