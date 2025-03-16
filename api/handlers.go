// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
	"github.com/nuklai/nuklaivm-external-subscriber/db/repository"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Returns API health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	}

	respondWithJSON(w, http.StatusOK, health)
}

// Returns a list of blocks
func (s *Server) getBlocksHandler(w http.ResponseWriter, r *http.Request) {
	limit, offset := getPaginationParams(r)

	startTime, endTime, err := getTimeRangeParams(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	blocks, err := s.repository.GetBlocksByTime(r.Context(), startTime, endTime, limit, offset)
	if err != nil {
		logger.Error("Failed to get blocks", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get blocks")
		return
	}

	response := map[string]interface{}{
		"blocks":     blocks,
		"pagination": map[string]int{"limit": limit, "offset": offset},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a block by it's height
func (s *Server) getBlockByHeightHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	heightStr := vars["height"]

	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid block height")
		return
	}

	block, err := s.repository.GetBlockByHeight(r.Context(), height)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Block not found")
		} else {
			logger.Error("Failed to get block by height", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to get block")
		}
		return
	}

	transactions, err := s.repository.GetTransactionsByBlockHeight(r.Context(), height)
	if err != nil {
		logger.Error("Failed to get transactions for block", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get transactions for block")
		return
	}

	response := map[string]interface{}{
		"block":        block,
		"transactions": transactions,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a list of transactions
func (s *Server) getTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	limit, offset := getPaginationParams(r)

	address := r.URL.Query().Get("address")

	var transactions []*models.Transaction
	var err error

	if address != "" {
		transactions, err = s.repository.GetTransactionsByAddress(r.Context(), address, limit, offset)
	} else {
		// TODO: Implement GetTransactions method in repo
		respondWithError(w, http.StatusNotImplemented, "Generic transaction listing not implemented yet")
		return
	}

	if err != nil {
		logger.Error("Failed to get transactions", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get transactions")
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"pagination":   map[string]int{"limit": limit, "offset": offset},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a transaction by it's hash
func (s *Server) getTransactionByHashHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	transaction, err := s.repository.GetTransactionByHash(r.Context(), hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Transaction not found")
		} else {
			logger.Error("Failed to get transaction by hash", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to get transaction")
		}
		return
	}

	events, err := s.repository.GetEventsByTxHash(r.Context(), hash)
	if err != nil {
		logger.Error("Failed to get events for transaction", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get events for transaction")
		return
	}

	response := map[string]interface{}{
		"transaction": transaction,
		"events":      events,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns account info
func (s *Server) getAccountHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	address := vars["address"]

	account, err := s.repository.GetAccountByAddress(r.Context(), address)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Account not found")
		} else {
			logger.Error("Failed to get account by address", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to get account")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, account)
}

// Returns transactions for an account
func (s *Server) getAccountTransactionsHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	address := vars["address"]

	limit, offset := getPaginationParams(r)

	transactions, err := s.repository.GetTransactionsByAddress(r.Context(), address, limit, offset)
	if err != nil {
		logger.Error("Failed to get transactions for account", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get transactions for account")
		return
	}

	response := map[string]interface{}{
		"address":      address,
		"transactions": transactions,
		"pagination":   map[string]int{"limit": limit, "offset": offset},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a list of datasets
func (s *Server) getDatasetsHandler(w http.ResponseWriter, r *http.Request) {

	limit, offset := getPaginationParams(r)

	datasets, err := s.repository.GetDatasets(r.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to get datasets", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get datasets")
		return
	}

	response := map[string]interface{}{
		"datasets":   datasets,
		"pagination": map[string]int{"limit": limit, "offset": offset},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a dataset by it's denom
func (s *Server) getDatasetByDenomHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	denom := vars["denom"]

	dataset, err := s.repository.GetDatasetByDenom(r.Context(), denom)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Dataset not found")
		} else {
			logger.Error("Failed to get dataset by denom", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to get dataset")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, dataset)
}

// Returns contributions for a dataset
func (s *Server) getDatasetContributionsHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	denom := vars["denom"]

	contributions, err := s.repository.GetDatasetContributions(r.Context(), denom)
	if err != nil {
		logger.Error("Failed to get contributions for dataset", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get contributions for dataset")
		return
	}

	response := map[string]interface{}{
		"denom":         denom,
		"contributions": contributions,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a list of validators
func (s *Server) getValidatorsHandler(w http.ResponseWriter, r *http.Request) {

	limit, offset := getPaginationParams(r)

	validators, err := s.repository.GetValidators(r.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to get validators", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get validators")
		return
	}

	response := map[string]interface{}{
		"validators": validators,
		"pagination": map[string]int{"limit": limit, "offset": offset},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a validator by address
func (s *Server) getValidatorByAddressHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	address := vars["address"]

	validator, err := s.repository.GetValidatorByOperatorAddress(r.Context(), address)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Try consensus address
			validator, err = s.repository.GetValidatorByConsensusAddress(r.Context(), address)
			if err != nil {
				respondWithError(w, http.StatusNotFound, "Validator not found")
				return
			}
		} else {
			logger.Error("Failed to get validator by address", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to get validator")
			return
		}
	}

	respondWithJSON(w, http.StatusOK, validator)
}

// Returns a list of events
func (s *Server) getEventsHandler(w http.ResponseWriter, r *http.Request) {

	limit, offset := getPaginationParams(r)

	eventType := r.URL.Query().Get("type")
	module := r.URL.Query().Get("module")

	var events []*models.Event
	var err error

	if eventType != "" {
		events, err = s.repository.GetEventsByType(r.Context(), eventType, limit, offset)
	} else if module != "" {
		events, err = s.repository.GetEventsByModule(r.Context(), module, limit, offset)
	} else {
		// TODO: Implement GetEvents method in repo
		respondWithError(w, http.StatusNotImplemented, "Generic event listing not implemented yet")
		return
	}

	if err != nil {
		logger.Error("Failed to get events", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get events")
		return
	}

	response := map[string]interface{}{
		"events":     events,
		"pagination": map[string]int{"limit": limit, "offset": offset},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Returns a list of event types
func (s *Server) getEventTypesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GetEventTypes method in repository
	respondWithError(w, http.StatusNotImplemented, "Event types endpoint not implemented yet")
}

// Helpers

// Extract limit and offset from query params
func getPaginationParams(r *http.Request) (int, int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default here
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default here
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Cap max limit
	if limit > 100 {
		limit = 100
	}

	return limit, offset
}

// Extract start_time and end_time from query params
func getTimeRangeParams(r *http.Request) (time.Time, time.Time, error) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var startTime time.Time
	var endTime time.Time
	var err error

	// Parse start time
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return startTime, endTime, fmt.Errorf("invalid start_time format, use RFC3339")
		}
	} else {
		// Default to 24 hours ago
		startTime = time.Now().UTC().Add(-24 * time.Hour)
	}

	// Parse end time
	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return startTime, endTime, fmt.Errorf("invalid end_time format, use RFC3339")
		}
	} else {
		// Default to now
		endTime = time.Now().UTC()
	}

	// Validate the time range
	if endTime.Before(startTime) {
		return startTime, endTime, fmt.Errorf("end_time must be after start_time")
	}

	return startTime, endTime, nil
}
