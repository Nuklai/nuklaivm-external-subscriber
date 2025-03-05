// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"time"
)

type Dataset struct {
	Denom              string `db:"denom" json:"denom"`
	Name               string `db:"name" json:"name"`
	Ticker             string `db:"ticker" json:"ticker"`
	Description        string `db:"description" json:"description"`
	URL                string `db:"url" json:"url"`
	Supply             string `db:"supply" json:"supply"`
	Categories         string `db:"categories" json:"categories"`
	LicenseName        string `db:"license_name" json:"license_name"`
	LicenseSymbol      string `db:"license_symbol" json:"license_symbol"`
	LicenseURL         string `db:"license_url" json:"license_url"`
	IsCommunityDataset bool   `db:"is_community_dataset" json:"is_community_dataset"`
	Metadata           string `db:"metadata" json:"metadata"`
	Owner              string `db:"owner" json:"owner"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// a user's contribution to a dataset
type DatasetContribution struct {
	ContributionID  string `db:"contribution_id" json:"contribution_id"`
	Denom           string `db:"denom" json:"denom"`
	DataContributor string `db:"data_contributor" json:"data_contributor"`
	DataLocation    string `db:"data_location" json:"data_location"`
	DataIdentifier  string `db:"data_identifier" json:"data_identifier"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// a pending contribution to a dataset
type PendingDatasetContribution struct {
	ContributionID  string `db:"contribution_id" json:"contribution_id"`
	Denom           string `db:"denom" json:"denom"`
	DataContributor string `db:"data_contributor" json:"data_contributor"`
	DataLocation    string `db:"data_location" json:"data_location"`
	DataIdentifier  string `db:"data_identifier" json:"data_identifier"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
