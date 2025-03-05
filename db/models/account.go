package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Account struct {
	Address     string    `db:"address" json:"address"`
	AccountType string    `db:"account_type" json:"account_type"`
	Balance     Coins     `db:"balance" json:"balance"`
	LastActive  time.Time `db:"last_active" json:"last_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Coin is a single denomination and its amount
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Coins []Coin

// Convert the Coins to a db value
func (c *Coins) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Convert a db value to coins
func (c *Coins) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, c)
}
