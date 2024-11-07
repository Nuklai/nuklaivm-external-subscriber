package api

import (
	"database/sql"
	"net/http"

	"github.com/nuklai/nuklaivm-external-subscriber/models"

	"github.com/gin-gonic/gin"
)

// GetAllBlocks retrieves all blocks with pagination
func GetAllBlocks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		blocks, err := models.FetchAllBlocks(db, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve blocks"})
			return
		}

		c.JSON(http.StatusOK, blocks)
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
