// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// GetAllAssets retrieves all assets
func GetAllAssets(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of assets
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM assets`).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count assets"})
			return
		}

		// Fetch paginated actions
		assets, err := models.FetchAllAssets(db, limit, offset)
		if err != nil {
			log.Printf("Error fetching assets: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve assets"})
			return
		}

		// Return response with counter
		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   assets,
		})
	}
}

// GetAssetsByType retrieves assets by their type
func GetAssetsByType(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		assetType := c.Param("type")
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of assets by type
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM assets WHERE asset_type_id = $1`, assetType).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count assets by type"})
			return
		}

		// Fetch paginated assets by type
		assets, err := models.FetchAssetsByType(db, assetType, limit, offset)
		if err != nil {
			log.Printf("Error fetching assets: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve assets by type"})
			return
		}

		// Return response with counter
		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   assets,
		})
	}
}

// GetAssetsByUser retrieves assets created by a specific user
func GetAssetsByUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.Param("user")
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get total count of assets created by user
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM assets WHERE asset_creator = $1`, user).Scan(&totalCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count assets for user"})
			return
		}

		// Fetch paginated assets by user
		assets, err := models.FetchAssetsByUser(db, user, limit, offset)
		if err != nil {
			log.Printf("Error fetching assets: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve assets for user"})
			return
		}

		// Return response with counter
		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   assets,
		})
	}
}
