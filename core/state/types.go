// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"context"
	"time"

	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
)

type StateRequest struct {
	Height             int64
	BlockTime          time.Time
	AccountAddresses   []string
	ValidatorAddresses []string
}

// The extracted state data
type StateResponse struct {
	Accounts      []*models.Account
	Validators    []*models.Validator
	Delegations   []*models.Delegation
	Datasets      []*models.Dataset
	WasmContracts []*models.WasmContract
	Proposals     []*models.Proposal
}

// IExtractor state extraction
type Extractor interface {
	ExtractState(ctx context.Context, req *StateRequest) (*StateResponse, error)
	ExtractAccounts(ctx context.Context, height int64, addresses []string) ([]*models.Account, error)
	ExtractValidators(ctx context.Context, height int64) ([]*models.Validator, error)
	ExtractDelegations(ctx context.Context, height int64, delegatorAddresses []string) ([]*models.Delegation, error)
	ExtractDatasets(ctx context.Context, height int64) ([]*models.Dataset, error)
	ExtractWasmContracts(ctx context.Context, height int64) ([]*models.WasmContract, error)
	ExtractProposals(ctx context.Context, height int64) ([]*models.Proposal, error)
}
