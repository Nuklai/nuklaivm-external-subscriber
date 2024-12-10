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

// GetAllAssets retrieves all assets with optional filters
func GetAllAssets(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		assetType := c.Query("type")
		user := c.Query("user")
		assetAddress := c.Query("asset_address")
		name := c.Query("name")
		symbol := c.Query("symbol")
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Normalize user to search with and without "0x" prefix
		if user != "" {
			user = strings.TrimPrefix(user, "0x")
		}
		if assetAddress != "" {
			assetAddress = strings.TrimPrefix(assetAddress, "0x")
		}

		// Get total count with filters
		totalCount, err := models.CountFilteredAssets(db, assetType, user, assetAddress, name, symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count assets"})
			return
		}

		// Fetch filtered assets with pagination
		assets, err := models.FetchFilteredAssets(db, assetType, user, assetAddress, name, symbol, limit, offset)
		if err != nil {
			log.Printf("Error fetching assets: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve assets"})
			return
		}

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

		user = strings.TrimPrefix(user, "0x")

		// Get total count of assets created by user
		var totalCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM assets WHERE asset_creator ILIKE $1`, "%"+user+"%").Scan(&totalCount)
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

// GetAssetByAddress retrieves a specific asset by its asset address
func GetAssetByAddress(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		assetAddress := c.Param("asset_address")

		assetAddress = strings.TrimPrefix(assetAddress, "0x")

		asset, err := models.FetchAssetByAddress(db, assetAddress)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
			return
		}

		c.JSON(http.StatusOK, asset)
	}
}
