package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// GetAllTransactions retrieves all transactions with pagination
func GetAllTransactions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		transactions, err := models.FetchAllTransactions(db, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions"})
			return
		}

		c.JSON(http.StatusOK, transactions)
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
