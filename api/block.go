package api

import (
	"database/sql"
	"net/http"

	"github.com/nuklai/nuklaivm-external-subscriber/models"

	"github.com/gin-gonic/gin"
)

// GetAllBlocks retrieves all blocks with pagination and total count
func GetAllBlocks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of blocks
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM blocks`).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count blocks"})
			return
		}

		// Fetch paginated blocks
		blocks, err := models.FetchAllBlocks(db, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve blocks"})
			return
		}

		// Return response with counter
		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   blocks,
		})
	}
}

func GetBlock(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		blockIdentifier := c.Param("identifier")

		block, err := models.FetchBlock(db, blockIdentifier)
		// Check for query errors
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
			return
		}

		c.JSON(http.StatusOK, block)
	}
}
