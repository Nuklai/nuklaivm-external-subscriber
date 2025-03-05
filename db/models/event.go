// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Event struct {
	ID          string     `db:"id" json:"id"`
	BlockHeight int64      `db:"block_height" json:"block_height"`
	BlockTime   time.Time  `db:"block_time" json:"block_time"`
	TxHash      string     `db:"tx_hash" json:"tx_hash"`
	TxIndex     int32      `db:"tx_index" json:"tx_index"`
	EventType   string     `db:"event_type" json:"event_type"`
	EventIndex  int32      `db:"event_index" json:"event_index"`
	Module      string     `db:"module" json:"module"`
	Attributes  Attributes `db:"attributes" json:"attributes"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// List of event attributes
type Attributes []Attribute

func (a *Attributes) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *Attributes) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, a)
}
