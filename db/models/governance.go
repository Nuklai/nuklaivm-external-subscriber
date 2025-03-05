// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Proposal struct {
	ProposalID      int64           `db:"proposal_id" json:"proposal_id"`
	Title           string          `db:"title" json:"title"`
	Description     string          `db:"description" json:"description"`
	ProposalType    string          `db:"proposal_type" json:"proposal_type"`
	ProposalStatus  string          `db:"proposal_status" json:"proposal_status"`
	SubmitTime      time.Time       `db:"submit_time" json:"submit_time"`
	DepositEndTime  time.Time       `db:"deposit_end_time" json:"deposit_end_time"`
	VotingStartTime time.Time       `db:"voting_start_time" json:"voting_start_time"`
	VotingEndTime   time.Time       `db:"voting_end_time" json:"voting_end_time"`
	Proposer        string          `db:"proposer" json:"proposer"`
	Content         ProposalContent `db:"content" json:"content"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// The content of a proposal
type ProposalContent json.RawMessage

func (pc *ProposalContent) Value() (driver.Value, error) {
	if pc == nil {
		return nil, nil
	}
	return json.Marshal(pc)
}

func (pc *ProposalContent) Scan(value interface{}) error {
	if value == nil {
		*pc = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, pc)
}

// Represents the deposit on a gov proposal
type ProposalDeposit struct {
	ID               string `db:"id" json:"id"`
	ProposalID       int64  `db:"proposal_id" json:"proposal_id"`
	DepositorAddress string `db:"depositor_address" json:"depositor_address"`
	Amount           Coins  `db:"amount" json:"amount"`
	BlockHeight      int64  `db:"block_height" json:"block_height"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Represents vote on a gov proposal
type ProposalVote struct {
	ID           string `db:"id" json:"id"`
	ProposalID   int64  `db:"proposal_id" json:"proposal_id"`
	VoterAddress string `db:"voter_address" json:"voter_address"`
	Option       string `db:"option" json:"option"`
	BlockHeight  int64  `db:"block_height" json:"block_height"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}