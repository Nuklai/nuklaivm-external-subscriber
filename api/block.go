// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/nuklai/nuklaivm-external-subscriber/models"

	"github.com/gin-gonic/gin"
)

// GetAllBlocks retrieves all blocks with pagination and total count
func GetAllBlocks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		blockHash := c.Query("block_hash")
		blockHeight := c.Query("block_height")
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Request a specific block
		if blockHash != "" || blockHeight != "" {
			block, err := models.FetchBlock(db, blockHeight, blockHash)
			if err != nil {
				log.Printf("Error fetching block: %v", err)
				c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
				return
			}
			c.JSON(http.StatusOK, block)
			return
		}

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
			log.Printf("Error fetching blocks: %v", err)
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

// func GetBlock(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		blockIdentifier := c.Param("identifier")

// 		block, err := models.FetchBlock(db, blockIdentifier)
// 		// Check for query errors
// 		if err != nil {
// 			log.Printf("Error fetching block: %v", err)
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, block)
// 	}
// }
