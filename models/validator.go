package models

import (
    "database/sql"
    "log"
)

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

// CountValidatorStakes gets total count of validator stakes
func CountValidatorStakes(db *sql.DB) (int, error) {
    var count int
    err := db.QueryRow(`SELECT COUNT(*) FROM validator_stake`).Scan(&count)
    if err != nil {
        log.Printf("Error counting validator stakes: %v", err)
        return 0, err
    }
    return count, nil
}

// FetchAllValidatorStakes retrieves validator stakes from the database with pagination
func FetchAllValidatorStakes(db *sql.DB, limit, offset string) ([]ValidatorStake, error) {
    rows, err := db.Query(`
        SELECT node_id, actor, stake_start_block, stake_end_block, 
               staked_amount, delegation_fee_rate, reward_address, 
               tx_hash, timestamp
        FROM validator_stake
        ORDER BY timestamp DESC
        LIMIT $1 OFFSET $2`, limit, offset)
    if err != nil {
        log.Printf("Database query error: %v", err)
        return nil, err
    }
    defer rows.Close()

    return scanValidatorStakes(rows)
}

// FetchValidatorStakeByNodeID retrieves a specific validator stake by node ID
func FetchValidatorStakeByNodeID(db *sql.DB, nodeID string) (ValidatorStake, error) {
    var stake ValidatorStake
    err := db.QueryRow(`
        SELECT node_id, actor, stake_start_block, stake_end_block, 
               staked_amount, delegation_fee_rate, reward_address, 
               tx_hash, timestamp
        FROM validator_stake
        WHERE node_id = $1`, nodeID).Scan(
            &stake.NodeID, &stake.Actor, &stake.StakeStartBlock,
            &stake.StakeEndBlock, &stake.StakedAmount,
            &stake.DelegationFeeRate, &stake.RewardAddress,
            &stake.TxHash, &stake.Timestamp,
    )
    if err != nil {
        log.Printf("Error fetching validator stake: %v", err)
        return stake, err
    }
    return stake, nil
}

// Helper function to scan validator stake rows
func scanValidatorStakes(rows *sql.Rows) ([]ValidatorStake, error) {
    var stakes []ValidatorStake
    for rows.Next() {
        var stake ValidatorStake
        if err := rows.Scan(
            &stake.NodeID, &stake.Actor, &stake.StakeStartBlock,
            &stake.StakeEndBlock, &stake.StakedAmount,
            &stake.DelegationFeeRate, &stake.RewardAddress,
            &stake.TxHash, &stake.Timestamp,
        ); err != nil {
            return nil, err
        }
        stakes = append(stakes, stake)
    }
    return stakes, rows.Err()
}