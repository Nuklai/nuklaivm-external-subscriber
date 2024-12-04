package models

import (
	"database/sql"
	"log"
)

type Asset struct {
	ID                           int    `json:"id"` // New field for the primary key
	AssetID                      string `json:"asset_id"`
	AssetTypeID                  int    `json:"asset_type_id"`
	AssetType                    string `json:"asset_type"`
	AssetCreator                 string `json:"asset_creator"`
	TxHash                       string `json:"tx_hash"`
	Name                         string `json:"name"`
	Symbol                       string `json:"symbol"`
	Decimals                     int    `json:"decimals"`
	Metadata                     string `json:"metadata"`
	MaxSupply                    int64  `json:"max_supply"`
	MintAdmin                    string `json:"mint_admin"`
	PauseUnpauseAdmin            string `json:"pause_unpause_admin"`
	FreezeUnfreezeAdmin          string `json:"freeze_unfreeze_admin"`
	EnableDisableKYCAccountAdmin string `json:"enable_disable_kyc_account_admin"`
	Timestamp                    string `json:"timestamp"`
}

func FetchAllAssets(db *sql.DB, limit, offset string) ([]Asset, error) {
	rows, err := db.Query(`SELECT * FROM assets ORDER BY timestamp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanAssets(rows)
}

func FetchAssetsByType(db *sql.DB, assetType, limit, offset string) ([]Asset, error) {
	rows, err := db.Query(`
		SELECT * FROM assets
		WHERE asset_type_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`, assetType, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanAssets(rows)
}

func FetchAssetsByUser(db *sql.DB, user, limit, offset string) ([]Asset, error) {
	rows, err := db.Query(`
		SELECT * FROM assets
		WHERE asset_creator = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`, user, limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanAssets(rows)
}

// Helper function to scan asset rows
func scanAssets(rows *sql.Rows) ([]Asset, error) {
	var assets []Asset

	for rows.Next() {
		var asset Asset
		if err := rows.Scan(&asset.ID, &asset.AssetID, &asset.AssetTypeID, &asset.AssetType, &asset.AssetCreator, &asset.TxHash, &asset.Name, &asset.Symbol, &asset.Decimals, &asset.Metadata,
			&asset.MaxSupply, &asset.MintAdmin, &asset.PauseUnpauseAdmin, &asset.FreezeUnfreezeAdmin, &asset.EnableDisableKYCAccountAdmin, &asset.Timestamp); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, rows.Err() // Check for errors during iteration
}
