package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

		rows, err := db.Query(`
            SELECT node_id, actor, stake_start_block, stake_end_block, staked_amount, delegation_fee_rate, reward_address, tx_hash, timestamp
            FROM validator_stake
            ORDER BY timestamp DESC
            LIMIT $1 OFFSET $2`, limit, offset)
		if err != nil {
			log.Printf("Error fetching validator stakes: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve validator stakes"})
			return
		}
		defer rows.Close()

		var stakes []ValidatorStake
		for rows.Next() {
			var stake ValidatorStake
			if err := rows.Scan(
				&stake.NodeID, &stake.Actor, &stake.StakeStartBlock,
				&stake.StakeEndBlock, &stake.StakedAmount,
				&stake.DelegationFeeRate, &stake.RewardAddress,
				&stake.TxHash, &stake.Timestamp,
			); err != nil {
				log.Printf("Error scanning stake row: %v\n", err)
				continue
			}
			stakes = append(stakes, stake)
		}

		c.JSON(http.StatusOK, gin.H{"items": stakes})
	}
}

// GetValidatorStakeByNodeID retrieves a specific validator stake by node ID
func GetValidatorStakeByNodeID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeID := c.Param("node_id")

		row := db.QueryRow(`
            SELECT node_id, actor, stake_start_block, stake_end_block, staked_amount, delegation_fee_rate, reward_address, tx_hash, timestamp
            FROM validator_stake
            WHERE node_id = $1`, nodeID)

		var stake ValidatorStake
		if err := row.Scan(
			&stake.NodeID, &stake.Actor, &stake.StakeStartBlock,
			&stake.StakeEndBlock, &stake.StakedAmount,
			&stake.DelegationFeeRate, &stake.RewardAddress,
			&stake.TxHash, &stake.Timestamp,
		); err != nil {
			log.Printf("Error fetching validator stake: %v\n", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Validator stake not found"})
			return
		}

		c.JSON(http.StatusOK, stake)
	}
}
