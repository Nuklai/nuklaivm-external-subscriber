// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Transaction struct {
	Hash        string    `db:"hash" json:"hash"`
	BlockHeight int64     `db:"block_height" json:"block_height"`
	BlockTime   time.Time `db:"block_time" json:"block_time"`
	Index       int32     `db:"index" json:"index"`
	GasWanted   int64     `db:"gas_wanted" json:"gas_wanted"`
	GasUsed     int64     `db:"gas_used" json:"gas_used"`
	Fee         string    `db:"fee" json:"fee"`
	Memo        string    `db:"memo" json:"memo"`
	ErrorLog    string    `db:"error_log" json:"error_log,omitempty"`
	Status      string    `db:"status" json:"status"`
	Codespace   string    `db:"codespace" json:"codespace,omitempty"`

	// JSON fields
	Messages Messages `db:"messages" json:"messages"`
	Signers  Signers  `db:"signers" json:"signers"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Messages []Message

// A transaction message
type Message struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

func (m *Messages) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *Messages) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, m)
}

// a list of transaction signers
type Signers []string

func (s *Signers) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *Signers) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, s)
}
