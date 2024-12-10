// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql"
	"fmt"
	"log"
)

type Asset struct {
	ID                           int    `json:"id"` // New field for the primary key
	AssetAddress                 string `json:"asset_address"`
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

// CountFilteredAssets counts assets based on optional filters
func CountFilteredAssets(db *sql.DB, assetType, user, assetAddress, name, symbol string) (int, error) {
	query, args := buildAssetFilterQuery("COUNT(*)", assetType, user, assetAddress, name, symbol)
	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// FetchFilteredAssets retrieves assets based on optional filters
func FetchFilteredAssets(db *sql.DB, assetType, user, assetAddress, name, symbol, limit, offset string) ([]Asset, error) {
	query, args := buildAssetFilterQuery("*", assetType, user, assetAddress, name, symbol)
	query = fmt.Sprintf("%s ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", query, len(args)+1, len(args)+2)

	args = append(args, limit, offset)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanAssets(rows)
}

// Helper function to construct filter queries for assets
func buildAssetFilterQuery(selectFields, assetType, user, assetAddress, name, symbol string) (string, []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM assets WHERE 1=1", selectFields)
	args := []interface{}{}
	argCounter := 1

	if assetType != "" {
		query += fmt.Sprintf(" AND asset_type_id = $%d", argCounter)
		args = append(args, assetType)
		argCounter++
	}

	if user != "" {
		query += fmt.Sprintf(" AND asset_creator ILIKE $%d", argCounter)
		args = append(args, "%"+user+"%")
		argCounter++
	}

	if assetAddress != "" {
		query += fmt.Sprintf(" AND asset_address ILIKE $%d", argCounter)
		args = append(args, "%"+assetAddress+"%")
		argCounter++
	}

	if name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argCounter)
		args = append(args, "%"+name+"%")
		argCounter++
	}

	if symbol != "" {
		query += fmt.Sprintf(" AND symbol ILIKE $%d", argCounter)
		args = append(args, "%"+symbol+"%")
		argCounter++
	}

	return query, args
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
		WHERE asset_creator ILIKE $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`, "%"+user+"%", limit, offset)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return scanAssets(rows)
}

// FetchAssetByAddress retrieves a specific asset by its asset address
func FetchAssetByAddress(db *sql.DB, assetAddress string) (Asset, error) {
	var asset Asset
	err := db.QueryRow(`SELECT * FROM assets WHERE asset_address ILIKE $1`, "%"+assetAddress+"%").Scan(
		&asset.ID, &asset.AssetAddress, &asset.AssetTypeID, &asset.AssetType, &asset.AssetCreator,
		&asset.TxHash, &asset.Name, &asset.Symbol, &asset.Decimals, &asset.Metadata,
		&asset.MaxSupply, &asset.MintAdmin, &asset.PauseUnpauseAdmin, &asset.FreezeUnfreezeAdmin,
		&asset.EnableDisableKYCAccountAdmin, &asset.Timestamp,
	)
	if err != nil {
		return asset, err
	}
	return asset, nil
}

// Helper function to scan asset rows
func scanAssets(rows *sql.Rows) ([]Asset, error) {
	var assets []Asset

	for rows.Next() {
		var asset Asset
		if err := rows.Scan(&asset.ID, &asset.AssetAddress, &asset.AssetTypeID, &asset.AssetType, &asset.AssetCreator, &asset.TxHash, &asset.Name, &asset.Symbol, &asset.Decimals, &asset.Metadata,
			&asset.MaxSupply, &asset.MintAdmin, &asset.PauseUnpauseAdmin, &asset.FreezeUnfreezeAdmin, &asset.EnableDisableKYCAccountAdmin, &asset.Timestamp); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, rows.Err() // Check for errors during iteration
}
