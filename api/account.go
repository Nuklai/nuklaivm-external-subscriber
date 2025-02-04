// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// GetAccountStats retrieves all account stats
func GetAccountStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := models.FetchAccountStats(db)
		if err != nil {
			log.Printf("Error fetching account stats: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve account stats"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// GetAccountDetails retrieves details for a specific account/address
func GetAccountDetails(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")

		details, err := models.FetchAccountByAddress(db, address)
		if err != nil {
			log.Printf("Error fetching account details: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}

		c.JSON(http.StatusOK, details)
	}
}

// GetAllAccounts retrieves accounts address, balance, transaction count
func GetAllAccounts(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "20")
		offset := c.DefaultQuery("offset", "0")

		// Get the total count of accounts
		totalCount, err := models.CountAccounts(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count accounts"})
			return
		}

		accounts, err := models.FetchAllAccounts(db, limit, offset)
		if err != nil {
			log.Printf("Error fetching accounts: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve accounts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   accounts,
		})
	}
}
