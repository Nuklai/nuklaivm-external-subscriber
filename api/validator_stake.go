package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
)

// Struct to represent a validator stake
type ValidatorStake struct {
	NodeID            string `json:"node_id"`
	Actor             string `json:"actor"`
	StakeStartBlock   int64  `json:"stake_start_block"`
	StakeEndBlock     int64  `json:"stake_end_block"`
	StakedAmount      int64  `json:"staked_amount"`
	DelegationFeeRate int64  `json:"delegation_fee_rate"`
	RewardAddress     string `json:"reward_address"`
	TxHash            string `json:"tx_hash"`
	Timestamp         string `json:"timestamp"`
}

// GetAllValidatorStakes retrieves all validator stakes with pagination
func GetAllValidatorStakes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")

		// Get validators total count
		totalCount, err := models.CountValidatorStakes(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to count validator stakes"})
			return
		}

		stakes, err := models.FetchAllValidatorStakes(db, limit, offset)
		if err != nil {
			log.Printf("Error fetching validator stakes: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve validator stakes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"counter": totalCount,
			"items":   stakes,
		})
	}
}

// GetValidatorStakeByNodeID retrieves a specific validator stake by node ID
func GetValidatorStakeByNodeID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeID := c.Param("node_id")

		stake, err := models.FetchValidatorStakeByNodeID(db, nodeID)
		if err != nil {
			log.Printf("Error fetching validator stake: %v\n", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Validator stake not found"})
			return
		}

		c.JSON(http.StatusOK, stake)
	}
}
