// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nuklai/nuklaivm-external-subscriber/client/rpc"
	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
	"github.com/nuklai/nuklaivm-external-subscriber/db/repository"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Implements IExtractor
type StateExtractor struct {
	rpcClient  *rpc.Client
	repository repository.Repository
}

func NewStateExtractor(rpcClient *rpc.Client, repo repository.Repository) Extractor {
	return &StateExtractor{
		rpcClient:  rpcClient,
		repository: repo,
	}
}

// Extract state data at the specified height
func (e *StateExtractor) ExtractState(ctx context.Context, req *StateRequest) (*StateResponse, error) {
	response := &StateResponse{}
	var err error

	// Extract accounts
	if len(req.AccountAddresses) > 0 {
		response.Accounts, err = e.ExtractAccounts(ctx, req.Height, req.AccountAddresses)
		if err != nil {
			return nil, fmt.Errorf("failed to extract accounts: %w", err)
		}
	}

	// Extract validators
	response.Validators, err = e.ExtractValidators(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to extract validators: %w", err)
	}

	// Extract delegations
	if len(req.AccountAddresses) > 0 {
		response.Delegations, err = e.ExtractDelegations(ctx, req.Height, req.AccountAddresses)
		if err != nil {
			return nil, fmt.Errorf("failed to extract delegations: %w", err)
		}
	}

	// Extract datasets
	response.Datasets, err = e.ExtractDatasets(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to extract datasets: %w", err)
	}

	// Extract WASM contracts
	response.WasmContracts, err = e.ExtractWasmContracts(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to extract WASM contracts: %w", err)
	}

	// Extract governance proposals
	response.Proposals, err = e.ExtractProposals(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to extract governance proposals: %w", err)
	}

	return response, nil
}

// Extract account info for the specified addresses
func (e *StateExtractor) ExtractAccounts(ctx context.Context, height int64, addresses []string) ([]*models.Account, error) {
	accounts := make([]*models.Account, 0, len(addresses))

	for _, addr := range addresses {
		path := fmt.Sprintf("cosmos/auth/v1beta1/accounts/%s", addr)
		resp, err := e.rpcClient.ABCIQuery(ctx, path, nil)
		if err != nil {
			logger.Warn("Failed to query account", zap.String("address", addr), zap.Error(err))
			continue
		}

		var authResponse struct {
			Account struct {
				Type    string `json:"@type"`
				Address string `json:"address"`
			} `json:"account"`
		}

		err = json.Unmarshal(resp.Response.Value, &authResponse)
		if err != nil {
			logger.Warn("Failed to unmarshal account response", zap.String("address", addr), zap.Error(err))
			continue
		}

		balancePath := fmt.Sprintf("cosmos/bank/v1beta1/balances/%s", addr)
		balanceResp, err := e.rpcClient.ABCIQuery(ctx, balancePath, nil)
		if err != nil {
			logger.Warn("Failed to query account balances", zap.String("address", addr), zap.Error(err))
			continue
		}

		var balanceResponse struct {
			Balances []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"balances"`
		}

		err = json.Unmarshal(balanceResp.Response.Value, &balanceResponse)
		if err != nil {
			logger.Warn("Failed to unmarshal balance response", zap.String("address", addr), zap.Error(err))
			continue
		}

		// Convert to Coins
		coins := make(models.Coins, 0, len(balanceResponse.Balances))
		for _, balance := range balanceResponse.Balances {
			coins = append(coins, models.Coin{
				Denom:  balance.Denom,
				Amount: balance.Amount,
			})
		}

		// Create account model
		account := &models.Account{
			Address:     addr,
			AccountType: authResponse.Account.Type,
			Balance:     coins,
			LastActive:  time.Now().UTC(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Extract validator info
func (e *StateExtractor) ExtractValidators(ctx context.Context, height int64) ([]*models.Validator, error) {
	// Query all validators availabe
	path := "cosmos/staking/v1beta1/validators"
	resp, err := e.rpcClient.ABCIQuery(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query validators: %w", err)
	}

	var validatorResponse struct {
		Validators []struct {
			OperatorAddress string `json:"operator_address"`
			ConsensusPubkey struct {
				Type string `json:"@type"`
				Key  string `json:"key"`
			} `json:"consensus_pubkey"`
			Status      string `json:"status"`
			Tokens      string `json:"tokens"`
			Description struct {
				Moniker         string `json:"moniker"`
				Identity        string `json:"identity"`
				Website         string `json:"website"`
				SecurityContact string `json:"security_contact"`
				Details         string `json:"details"`
			} `json:"description"`
			Commission struct {
				CommissionRates struct {
					Rate string `json:"rate"`
				} `json:"commission_rates"`
			} `json:"commission"`
		} `json:"validators"`
	}

	err = json.Unmarshal(resp.Response.Value, &validatorResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator response: %w", err)
	}

	validators := make([]*models.Validator, 0, len(validatorResponse.Validators))

	for _, v := range validatorResponse.Validators {
		// Decode consensus pubkey to get consensus address
		pubkeyBytes, err := base64.StdEncoding.DecodeString(v.ConsensusPubkey.Key)
		if err != nil {
			logger.Warn("Failed to decode validator pubkey",
				zap.String("operator_address", v.OperatorAddress),
				zap.Error(err))
			continue
		}

		// TODO: Convert pubkey to consensus address
		consensusAddr := fmt.Sprintf("%X", pubkeyBytes[:20])

		// Create validator model
		validator := &models.Validator{
			ConsensusAddress: consensusAddr,
			OperatorAddress:  v.OperatorAddress,
			AccountAddress:   "", // TODO: Convert operator address to account address
			Moniker:          v.Description.Moniker,
			Identity:         v.Description.Identity,
			Website:          v.Description.Website,
			SecurityContact:  v.Description.SecurityContact,
			Details:          v.Description.Details,
			CommissionRate:   v.Commission.CommissionRates.Rate,
			VotingPower:      v.Tokens,
			Status:           v.Status,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}

		validators = append(validators, validator)
	}

	return validators, nil
}

// Extracts delegation info for the specified delegator addresses
func (e *StateExtractor) ExtractDelegations(ctx context.Context, height int64, delegatorAddresses []string) ([]*models.Delegation, error) {
	delegations := make([]*models.Delegation, 0)

	for _, addr := range delegatorAddresses {
		// Query delegations availabe
		path := fmt.Sprintf("cosmos/staking/v1beta1/delegations/%s", addr)
		resp, err := e.rpcClient.ABCIQuery(ctx, path, nil)
		if err != nil {
			logger.Warn("Failed to query delegations", zap.String("address", addr), zap.Error(err))
			continue
		}

		var delegationResponse struct {
			DelegationResponses []struct {
				Delegation struct {
					DelegatorAddress string `json:"delegator_address"`
					ValidatorAddress string `json:"validator_address"`
					Shares           string `json:"shares"`
				} `json:"delegation"`
				Balance struct {
					Denom  string `json:"denom"`
					Amount string `json:"amount"`
				} `json:"balance"`
			} `json:"delegation_responses"`
		}

		err = json.Unmarshal(resp.Response.Value, &delegationResponse)
		if err != nil {
			logger.Warn("Failed to unmarshal delegation response", zap.String("address", addr), zap.Error(err))
			continue
		}

		for _, d := range delegationResponse.DelegationResponses {
			delegation := &models.Delegation{
				ID:               fmt.Sprintf("%s-%s", d.Delegation.DelegatorAddress, d.Delegation.ValidatorAddress),
				DelegatorAddress: d.Delegation.DelegatorAddress,
				ValidatorAddress: d.Delegation.ValidatorAddress,
				Amount:           d.Balance.Amount,
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
			}
			delegations = append(delegations, delegation)
		}
	}

	return delegations, nil
}

// Extract dataset info
func (e *StateExtractor) ExtractDatasets(ctx context.Context, height int64) ([]*models.Dataset, error) {
	// Query all datasets availabe
	path := "nuklai/dataset/v1/datasets"
	resp, err := e.rpcClient.ABCIQuery(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query datasets: %w", err)
	}

	var datasetResponse struct {
		Datasets []struct {
			Denom              string `json:"denom"`
			Name               string `json:"name"`
			Ticker             string `json:"ticker"`
			Description        string `json:"description"`
			URL                string `json:"url"`
			Supply             string `json:"supply"`
			Categories         string `json:"categories"`
			LicenseName        string `json:"license_name"`
			LicenseSymbol      string `json:"license_symbol"`
			LicenseURL         string `json:"license_url"`
			IsCommunityDataset bool   `json:"is_community_dataset"`
			Metadata           string `json:"metadata"`
			Owner              string `json:"owner"`
		} `json:"datasets"`
	}

	err = json.Unmarshal(resp.Response.Value, &datasetResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal dataset response: %w", err)
	}

	datasets := make([]*models.Dataset, 0, len(datasetResponse.Datasets))

	for _, d := range datasetResponse.Datasets {
		dataset := &models.Dataset{
			Denom:              d.Denom,
			Name:               d.Name,
			Ticker:             d.Ticker,
			Description:        d.Description,
			URL:                d.URL,
			Supply:             d.Supply,
			Categories:         d.Categories,
			LicenseName:        d.LicenseName,
			LicenseSymbol:      d.LicenseSymbol,
			LicenseURL:         d.LicenseURL,
			IsCommunityDataset: d.IsCommunityDataset,
			Metadata:           d.Metadata,
			Owner:              d.Owner,
			CreatedAt:          time.Now().UTC(),
			UpdatedAt:          time.Now().UTC(),
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

// Extracts WASM contract info
func (e *StateExtractor) ExtractWasmContracts(ctx context.Context, height int64) ([]*models.WasmContract, error) {
	// Query all contracts available
	path := "cosmwasm/wasm/v1/contracts"
	resp, err := e.rpcClient.ABCIQuery(ctx, path, nil)
	if err != nil {
		logger.Warn("Failed to query WASM contracts, WASM module might not be enabled", zap.Error(err))
		return []*models.WasmContract{}, nil
	}

	var contractResponse struct {
		Contracts []string `json:"contracts"`
	}

	err = json.Unmarshal(resp.Response.Value, &contractResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract response: %w", err)
	}

	contracts := make([]*models.WasmContract, 0, len(contractResponse.Contracts))

	for _, contractAddr := range contractResponse.Contracts {
		// Query contract info
		contractPath := fmt.Sprintf("cosmwasm/wasm/v1/contract/%s", contractAddr)
		contractResp, err := e.rpcClient.ABCIQuery(ctx, contractPath, nil)
		if err != nil {
			logger.Warn("Failed to query contract info", zap.String("address", contractAddr), zap.Error(err))
			continue
		}

		var contractInfo struct {
			Address   string          `json:"address"`
			CodeID    string          `json:"code_id"`
			Creator   string          `json:"creator"`
			Admin     string          `json:"admin"`
			Label     string          `json:"label"`
			CreatedAt json.RawMessage `json:"created"`
			IBCPortID string          `json:"ibc_port_id"`
			InitMsg   json.RawMessage `json:"init_msg"`
		}

		err = json.Unmarshal(contractResp.Response.Value, &contractInfo)
		if err != nil {
			logger.Warn("Failed to unmarshal contract info", zap.String("address", contractAddr), zap.Error(err))
			continue
		}

		// Parse code ID as int64
		var codeID int64
		_, err = fmt.Sscanf(contractInfo.CodeID, "%d", &codeID)
		if err != nil {
			logger.Warn("Failed to parse code ID", zap.String("code_id", contractInfo.CodeID), zap.Error(err))
			continue
		}

		contract := &models.WasmContract{
			Address:           contractAddr,
			CodeID:            codeID,
			Creator:           contractInfo.Creator,
			Admin:             contractInfo.Admin,
			Label:             contractInfo.Label,
			CreatedHeight:     height, // Use current height as the fallback
			InstantiateTxHash: "",     // Hmmmm, cannot get this from ABCI query
			InitMsg:           contractInfo.InitMsg,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// Extract governance proposals
func (e *StateExtractor) ExtractProposals(ctx context.Context, height int64) ([]*models.Proposal, error) {
	// Query all proposals available
	path := "cosmos/gov/v1/proposals"
	resp, err := e.rpcClient.ABCIQuery(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query proposals: %w", err)
	}

	var proposalResponse struct {
		Proposals []struct {
			ID               string          `json:"id"`
			Messages         json.RawMessage `json:"messages"`
			Status           string          `json:"status"`
			FinalTallyResult struct {
				YesCount        string `json:"yes_count"`
				NoCount         string `json:"no_count"`
				AbstainCount    string `json:"abstain_count"`
				NoWithVetoCount string `json:"no_with_veto_count"`
			} `json:"final_tally_result"`
			SubmitTime      time.Time `json:"submit_time"`
			DepositEndTime  time.Time `json:"deposit_end_time"`
			VotingStartTime time.Time `json:"voting_start_time"`
			VotingEndTime   time.Time `json:"voting_end_time"`
			Metadata        string    `json:"metadata"`
			Title           string    `json:"title"`
			Summary         string    `json:"summary"`
			Proposer        string    `json:"proposer"`
		} `json:"proposals"`
	}

	err = json.Unmarshal(resp.Response.Value, &proposalResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal proposal response: %w", err)
	}

	proposals := make([]*models.Proposal, 0, len(proposalResponse.Proposals))

	for _, p := range proposalResponse.Proposals {
		var proposalID int64
		_, err = fmt.Sscanf(p.ID, "%d", &proposalID)
		if err != nil {
			logger.Warn("Failed to parse proposal ID", zap.String("id", p.ID), zap.Error(err))
			continue
		}

		content := models.ProposalContent(p.Messages)

		proposal := &models.Proposal{
			ProposalID:      proposalID,
			Title:           p.Title,
			Description:     p.Summary,
			ProposalType:    "text",
			ProposalStatus:  p.Status,
			SubmitTime:      p.SubmitTime,
			DepositEndTime:  p.DepositEndTime,
			VotingStartTime: p.VotingStartTime,
			VotingEndTime:   p.VotingEndTime,
			Proposer:        p.Proposer,
			Content:         content,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		proposals = append(proposals, proposal)
	}

	return proposals, nil
}
