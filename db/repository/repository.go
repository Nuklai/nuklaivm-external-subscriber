// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package repository

import (
	"context"
	"time"

	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
)

// Define the interface for db ops
type Repository interface {
	// Blocks
	SaveBlock(ctx context.Context, block *models.Block) error
	GetBlockByHeight(ctx context.Context, height int64) (*models.Block, error)
	GetBlockByHash(ctx context.Context, hash string) (*models.Block, error)
	GetLatestBlock(ctx context.Context) (*models.Block, error)
	GetBlocksInRange(ctx context.Context, startHeight, endHeight int64) ([]*models.Block, error)
	GetBlocksByTime(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.Block, error)

	// Transactions
	SaveTransaction(ctx context.Context, tx *models.Transaction) error
	GetTransactionByHash(ctx context.Context, hash string) (*models.Transaction, error)
	GetTransactionsByBlockHeight(ctx context.Context, height int64) ([]*models.Transaction, error)
	GetTransactionsByAddress(ctx context.Context, address string, limit, offset int) ([]*models.Transaction, error)
	GetTransactions(ctx context.Context, limit, offset int) ([]*models.Transaction, error)

	// Events
	SaveEvent(ctx context.Context, event *models.Event) error
	GetEventsByBlockHeight(ctx context.Context, height int64) ([]*models.Event, error)
	GetEventsByTxHash(ctx context.Context, txHash string) ([]*models.Event, error)
	GetEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*models.Event, error)
	GetEventsByModule(ctx context.Context, module string, limit, offset int) ([]*models.Event, error)

	// Accounts
	SaveAccount(ctx context.Context, account *models.Account) error
	GetAccountByAddress(ctx context.Context, address string) (*models.Account, error)
	UpdateAccountBalance(ctx context.Context, address string, balance models.Coins) error

	// Datasets
	SaveDataset(ctx context.Context, dataset *models.Dataset) error
	GetDatasetByDenom(ctx context.Context, denom string) (*models.Dataset, error)
	GetDatasets(ctx context.Context, limit, offset int) ([]*models.Dataset, error)
	UpdateDataset(ctx context.Context, dataset *models.Dataset) error
	SaveDatasetContribution(ctx context.Context, contribution *models.DatasetContribution) error
	GetDatasetContributions(ctx context.Context, denom string) ([]*models.DatasetContribution, error)
	SavePendingDatasetContribution(ctx context.Context, contribution *models.PendingDatasetContribution) error
	GetPendingDatasetContributions(ctx context.Context, denom string) ([]*models.PendingDatasetContribution, error)
	DeletePendingDatasetContribution(ctx context.Context, contributionID string) error

	// Validators
	SaveValidator(ctx context.Context, validator *models.Validator) error
	GetValidatorByConsensusAddress(ctx context.Context, consensusAddress string) (*models.Validator, error)
	GetValidatorByOperatorAddress(ctx context.Context, operatorAddress string) (*models.Validator, error)
	GetValidators(ctx context.Context, limit, offset int) ([]*models.Validator, error)
	UpdateValidator(ctx context.Context, validator *models.Validator) error
	SaveValidatorUptime(ctx context.Context, uptime *models.ValidatorUptime) error
	GetValidatorUptime(ctx context.Context, consensusAddress string, startHeight, endHeight int64) ([]*models.ValidatorUptime, error)

	// Validator rewards
	SaveValidatorReward(ctx context.Context, reward *models.ValidatorReward) error
	GetValidatorRewards(ctx context.Context, validatorAddress string, startHeight, endHeight int64) ([]*models.ValidatorReward, error)
	GetDelegatorRewards(ctx context.Context, delegatorAddress string, startHeight, endHeight int64) ([]*models.ValidatorReward, error)

	// Validator reward slashing
	SaveSlashingEvent(ctx context.Context, slashing *models.SlashingEvent) error
	GetSlashingEvents(ctx context.Context, validatorAddress string, startHeight, endHeight int64) ([]*models.SlashingEvent, error)

	// Delegations
	SaveDelegation(ctx context.Context, delegation *models.Delegation) error
	GetDelegationsByDelegator(ctx context.Context, delegatorAddress string) ([]*models.Delegation, error)
	GetDelegationsByValidator(ctx context.Context, validatorAddress string) ([]*models.Delegation, error)
	UpdateDelegation(ctx context.Context, delegation *models.Delegation) error

	// Governance
	SaveProposal(ctx context.Context, proposal *models.Proposal) error
	GetProposalByID(ctx context.Context, proposalID int64) (*models.Proposal, error)
	GetProposals(ctx context.Context, limit, offset int) ([]*models.Proposal, error)
	UpdateProposal(ctx context.Context, proposal *models.Proposal) error
	SaveProposalDeposit(ctx context.Context, deposit *models.ProposalDeposit) error
	GetProposalDeposits(ctx context.Context, proposalID int64) ([]*models.ProposalDeposit, error)
	SaveProposalVote(ctx context.Context, vote *models.ProposalVote) error
	GetProposalVotes(ctx context.Context, proposalID int64, limit, offset int) ([]*models.ProposalVote, error)

	// WASM
	SaveWasmCode(ctx context.Context, code *models.WasmCode) error
	GetWasmCodeByID(ctx context.Context, codeID int64) (*models.WasmCode, error)
	GetWasmCodes(ctx context.Context, limit, offset int) ([]*models.WasmCode, error)
	SaveWasmContract(ctx context.Context, contract *models.WasmContract) error
	GetWasmContractByAddress(ctx context.Context, address string) (*models.WasmContract, error)
	GetWasmContractsByCodeID(ctx context.Context, codeID int64) ([]*models.WasmContract, error)
	UpdateWasmContract(ctx context.Context, contract *models.WasmContract) error
	SaveWasmExecute(ctx context.Context, execute *models.WasmExecute) error
	GetWasmExecutesByContract(ctx context.Context, contractAddress string, limit, offset int) ([]*models.WasmExecute, error)
	SaveWasmQuery(ctx context.Context, query *models.WasmQuery) error
	GetWasmQueriesByContract(ctx context.Context, contractAddress string, limit, offset int) ([]*models.WasmQuery, error)

	// IBC
	SaveIBCChannel(ctx context.Context, channel *models.IBCChannel) error
	GetIBCChannelByID(ctx context.Context, channelID, portID string) (*models.IBCChannel, error)
	GetIBCChannels(ctx context.Context, limit, offset int) ([]*models.IBCChannel, error)
	UpdateIBCChannel(ctx context.Context, channel *models.IBCChannel) error
	SaveIBCConnection(ctx context.Context, connection *models.IBCConnection) error
	GetIBCConnectionByID(ctx context.Context, connectionID string) (*models.IBCConnection, error)
	GetIBCConnections(ctx context.Context, limit, offset int) ([]*models.IBCConnection, error)
	UpdateIBCConnection(ctx context.Context, connection *models.IBCConnection) error
	SaveIBCClient(ctx context.Context, client *models.IBCClient) error
	GetIBCClientByID(ctx context.Context, clientID string) (*models.IBCClient, error)
	GetIBCClients(ctx context.Context, limit, offset int) ([]*models.IBCClient, error)
	UpdateIBCClient(ctx context.Context, client *models.IBCClient) error
	SaveIBCPacket(ctx context.Context, packet *models.IBCPacket) error
	GetIBCPacketByID(ctx context.Context, id string) (*models.IBCPacket, error)
	GetIBCPacketsByChannel(ctx context.Context, channelID, portID string, limit, offset int) ([]*models.IBCPacket, error)
	UpdateIBCPacket(ctx context.Context, packet *models.IBCPacket) error
}
