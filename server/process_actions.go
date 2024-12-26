// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package server

import (
	"database/sql"
)

func processCreateAssetID(dbConn *sql.DB, actionInput map[string]interface{}, actionOutput map[string]interface{}, sponsor, txID, timestamp string) error {
	assetID := actionOutput["asset_address"].(string)
	assetTypeID := actionInput["asset_type"].(float64)
	assetType := map[float64]string{0: "fungible", 1: "non-fungible", 2: "fractional"}[assetTypeID]

	// Insert asset into the assets table
	_, err := dbConn.Exec(`
        INSERT INTO assets (
            asset_address, asset_type_id, asset_type, asset_creator, tx_hash, name, symbol, decimals, metadata, max_supply, mint_admin, pause_unpause_admin, freeze_unfreeze_admin, enable_disable_kyc_account_admin, timestamp
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        ON CONFLICT (asset_address) DO UPDATE
        SET asset_type_id = EXCLUDED.asset_type_id,
            asset_type = EXCLUDED.asset_type,
            asset_creator = EXCLUDED.asset_creator,
            tx_hash = EXCLUDED.tx_hash,
            name = EXCLUDED.name,
            symbol = EXCLUDED.symbol,
            decimals = EXCLUDED.decimals,
            metadata = EXCLUDED.metadata,
            max_supply = EXCLUDED.max_supply,
            mint_admin = EXCLUDED.mint_admin,
            pause_unpause_admin = EXCLUDED.pause_unpause_admin,
            freeze_unfreeze_admin = EXCLUDED.freeze_unfreeze_admin,
            enable_disable_kyc_account_admin = EXCLUDED.enable_disable_kyc_account_admin,
            timestamp = EXCLUDED.timestamp
    `, assetID, assetTypeID, assetType, sponsor, txID,
		actionInput["name"], actionInput["symbol"], actionInput["decimals"],
		actionInput["metadata"], actionInput["max_supply"], actionInput["mint_admin"],
		actionInput["pause_unpause_admin"], actionInput["freeze_unfreeze_admin"],
		actionInput["enable_disable_kyc_account_admin"], timestamp)
	return err
}

func processRegisterValidatorStakeID(dbConn *sql.DB, actionOutput map[string]interface{}, sponsor, txID, timestamp string) error {
	// Parse the action input
	nodeID := actionOutput["node_id"].(string)
	stakeStartBlock := uint64(actionOutput["stake_start_block"].(float64))
	stakeEndBlock := uint64(actionOutput["stake_end_block"].(float64))
	stakedAmount := uint64(actionOutput["staked_amount"].(float64))
	delegationFeeRate := uint64(actionOutput["delegation_fee_rate"].(float64))
	rewardAddress := actionOutput["reward_address"].(string)

	// Save the validator stake in the database
	_, err := dbConn.Exec(`
            INSERT INTO validator_stake (
                node_id, actor, stake_start_block, stake_end_block, staked_amount, delegation_fee_rate, reward_address, tx_hash, timestamp
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            ON CONFLICT (node_id, stake_start_block) DO UPDATE
            SET stake_end_block = EXCLUDED.stake_end_block,
                staked_amount = EXCLUDED.staked_amount,
                delegation_fee_rate = EXCLUDED.delegation_fee_rate,
                reward_address = EXCLUDED.reward_address,
                tx_hash = EXCLUDED.tx_hash,
                timestamp = EXCLUDED.timestamp`,
		nodeID, sponsor, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, rewardAddress, txID, timestamp,
	)
	return err
}
