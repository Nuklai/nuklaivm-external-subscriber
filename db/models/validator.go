package models

import (
	"time"
)

type Validator struct {
	ConsensusAddress string `db:"consensus_address" json:"consensus_address"`
	OperatorAddress  string `db:"operator_address" json:"operator_address"`
	AccountAddress   string `db:"account_address" json:"account_address"`
	Moniker          string `db:"moniker" json:"moniker"`
	Identity         string `db:"identity" json:"identity"`
	Website          string `db:"website" json:"website"`
	SecurityContact  string `db:"security_contact" json:"security_contact"`
	Details          string `db:"details" json:"details"`
	CommissionRate   string `db:"commission_rate" json:"commission_rate"`
	VotingPower      string `db:"voting_power" json:"voting_power"`
	Status           string `db:"status" json:"status"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Tracks validator uptime
type ValidatorUptime struct {
	ID               string    `db:"id" json:"id"`
	ConsensusAddress string    `db:"consensus_address" json:"consensus_address"`
	BlockHeight      int64     `db:"block_height" json:"block_height"`
	BlockTime        time.Time `db:"block_time" json:"block_time"`
	Signed           bool      `db:"signed" json:"signed"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Track delegations from an account to a validator
type Delegation struct {
	ID               string `db:"id" json:"id"`
	DelegatorAddress string `db:"delegator_address" json:"delegator_address"`
	ValidatorAddress string `db:"validator_address" json:"validator_address"`
	Amount           string `db:"amount" json:"amount"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
