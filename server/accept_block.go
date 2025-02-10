// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
	"github.com/lib/pq"
	"github.com/nuklai/nuklaivm-external-subscriber/consts"
	"github.com/nuklai/nuklaivm-external-subscriber/db"
	vmconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/vm"
)

// handleAcceptBlock processes a new block
func handleAcceptBlock(dbConn *sql.DB, parser chain.Parser, req *pb.BlockRequest) error {
	if parser == nil {
		log.Println("Parser is not initialized. Rejecting the request.")
		return errors.New("parser not initialized")
	}

	blockData := req.GetBlockData()

	executedBlock, err := chain.UnmarshalExecutedBlock(blockData, parser)
	if err != nil {
		log.Printf("Error parsing block data: %v\n", err)
		return err
	}

	blk := executedBlock.Block
	blockHeight := blk.Hght

	if blockHeight == 1 {
		log.Println("First block detected (genesis). Resetting the database...")

		// Drop all tables
		_, err := dbConn.Exec(`
			DROP TABLE IF EXISTS blocks, transactions, actions, assets CASCADE;
		`)
		if err != nil {
			log.Printf("Error dropping existing tables: %v\n", err)
			return err
		}

		// Re-create the schema
		err = db.CreateSchema(dbConn)
		if err != nil {
			log.Printf("Error re-creating schema: %v\n", err)
			return err
		}
		log.Println("Database reset and schema re-created successfully.")
	}

	err = processBlockData(dbConn, executedBlock)
	if err != nil {
		log.Printf("Error processing block data: %v\n", err)
		return err
	}

	return nil
}

// processBlockData saves block data to the database
func processBlockData(dbConn *sql.DB, executedBlock *chain.ExecutedBlock) error {
	blk := executedBlock.Block
	blockHash := executedBlock.BlockID.String()
	blockHeight := blk.Hght
	parentHash := blk.Prnt.String()
	stateRoot := blk.StateRoot.String()
	timestamp := time.UnixMilli(blk.Tmstmp).UTC().Format(time.RFC3339)
	blockSize := blk.Size()
	txCount := len(blk.Txs)
	avgTxSize := 0.0
	if txCount > 0 {
		avgTxSize = float64(blockSize) / float64(txCount)
	}

	log.Printf("Block Details: Height: %d, Hash: %s, ParentHash: %s, Transactions: %d\n", blockHeight, blockHash, parentHash, len(blk.Txs))

	uniqueParticipants := make(map[string]struct{})
	totalFee := uint64(0)

	for i, tx := range blk.Txs {
		txID := tx.ID().String()
		sponsor := tx.Sponsor().String()
		fee := uint64(0)
		outputs := []map[string]interface{}{}
		success := false

		actions := []map[string]interface{}{}
		actors := make(map[string]struct{})
		receivers := make(map[string]struct{})

		if i < len(executedBlock.Results) {
			result := executedBlock.Results[i]
			fee = result.Fee
			success = result.Success

			if success {
				// Parse outputs if available
				for _, outputBytes := range result.Outputs {
					packer := codec.NewReader(outputBytes, len(outputBytes))
					r, err := vm.OutputParser.Unmarshal(packer)
					if err == nil {
						outputJSON, err := json.Marshal(r)
						if err == nil {
							var outputMap map[string]interface{}
							json.Unmarshal(outputJSON, &outputMap)
							outputs = append(outputs, outputMap)

							// Add actor and receiver to uniqueParticipants and individual maps
							if actor, ok := outputMap["actor"].(string); ok && actor != "" {
								uniqueParticipants[actor] = struct{}{}
								actors[actor] = struct{}{}
							}
							if receiver, ok := outputMap["receiver"].(string); ok && receiver != "" {
								uniqueParticipants[receiver] = struct{}{}
								receivers[receiver] = struct{}{}
							}
						}
					}
				}
			}
		}
		totalFee += fee
		uniqueParticipants[sponsor] = struct{}{}

		log.Printf("\tTransaction %d: %s\n", i+1, txID)
		log.Printf("\tOutputs: %v\n", outputs)

		// Process and aggregate actions for the transaction
		for j, action := range tx.Actions {
			actionType := action.GetTypeID()
			actionName, ok := consts.ActionNames[actionType]
			if !ok {
				actionName = "Unknown"
			}

			actionInputJSON := "{}"
			if inputDetails, err := json.Marshal(action); err != nil {
				log.Printf("Error marshaling action input: %v\n", err)
			} else {
				actionInputJSON = string(inputDetails)
			}

			actionOutputsJSON := "{}"
			if j < len(outputs) {
				actionOutputs := outputs[j]
				if actionOutputs != nil {
					actionOutputsBytes, err := json.Marshal(actionOutputs)
					if err != nil {
						log.Printf("Error marshaling action outputs: %v\n", err)
					} else {
						actionOutputsJSON = string(actionOutputsBytes)
					}
				}
			}

			actionEntry := map[string]interface{}{
				"ActionTypeID": actionType,
				"ActionType":   actionName,
				"Input":        json.RawMessage(actionInputJSON),
				"Output":       json.RawMessage(actionOutputsJSON),
			}
			actions = append(actions, actionEntry)

			log.Printf("\t\tAction %d: Type: %d, Input: %s, Output: %s\n", j+1, actionType, actionInputJSON, actionOutputsJSON)

			// Save the action in the actions table
			_, err := dbConn.Exec(`
				INSERT INTO actions (tx_hash, action_type, action_name, action_index, input, output, timestamp)
				VALUES ($1, $2, $3, $4, $5::json, $6::json, $7)
				ON CONFLICT (tx_hash, action_type, action_index) DO UPDATE
				SET input = EXCLUDED.input,
						output = EXCLUDED.output,
						timestamp = EXCLUDED.timestamp`,
				txID, actionType, actionName, j, actionInputJSON, actionOutputsJSON, timestamp)
			if err != nil {
				log.Printf("Error saving action to database: %v\n", err)
			}

			// Update the action total
			if err := updateActionVolume(dbConn, actionType, actionName); err != nil {
				log.Printf("Error updating action total: %v\n", err)
			}

			// Handle special actions
			// Parse actionInputJSON into map[string]interface{}
			var actionInput map[string]interface{}
			err = json.Unmarshal([]byte(actionInputJSON), &actionInput)
			if err != nil {
				log.Printf("Error unmarshaling action input: %v\n", err)
				continue
			}
			actionOutput := outputs[j]
			var actionError error
			switch actionType {
			case vmconsts.CreateAssetID:
				actionError = processCreateAssetID(dbConn, actionInput, actionOutput, sponsor, txID, timestamp)
			case vmconsts.RegisterValidatorStakeID:
				actionError = processRegisterValidatorStakeID(dbConn, actionOutput, sponsor, txID, timestamp)
			}
			if actionError != nil {
				log.Printf("Error processing action '%s': %v\n", actionName, actionError)
				return actionError
			}
		}

		// Convert actions to JSON for storing in the transactions table
		actionsJSON, err := json.Marshal(actions)
		if err != nil {
			log.Printf("Error marshaling actions: %v\n", err)
			continue
		}

		// Convert actors and receivers to slices of strings
		actorsSlice := getKeysFromMap(actors)
		receiversSlice := getKeysFromMap(receivers)

		// Save the transaction with aggregated actions
		_, err = dbConn.Exec(`
    INSERT INTO transactions (tx_hash, block_hash, sponsor, actors, receivers, max_fee, success, fee, actions, timestamp)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::json, $10)
    ON CONFLICT (tx_hash) DO UPDATE
    SET block_hash = EXCLUDED.block_hash,
        sponsor = EXCLUDED.sponsor,
        actors = EXCLUDED.actors,
        receivers = EXCLUDED.receivers,
        max_fee = EXCLUDED.max_fee,
        success = EXCLUDED.success,
        fee = EXCLUDED.fee,
        actions = EXCLUDED.actions,
        timestamp = EXCLUDED.timestamp`,
			txID, blockHash, sponsor, pq.Array(actorsSlice), pq.Array(receiversSlice),
			tx.MaxFee(), success, fee, actionsJSON, timestamp)
		if err != nil {
			log.Printf("Error saving transaction to database: %v\n", err)
		}
	}

	// Save the new block data to the database
	_, err := dbConn.Exec(`
        INSERT INTO blocks (block_height, block_hash, parent_block_hash, state_root, block_size, tx_count, total_fee, avg_tx_size, unique_participants, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (block_height) DO UPDATE
        SET block_hash = EXCLUDED.block_hash,
            parent_block_hash = EXCLUDED.parent_block_hash,
            state_root = EXCLUDED.state_root,
            block_size = EXCLUDED.block_size,
            tx_count = EXCLUDED.tx_count,
            total_fee = EXCLUDED.total_fee,
            avg_tx_size = EXCLUDED.avg_tx_size,
            unique_participants = EXCLUDED.unique_participants,
            timestamp = EXCLUDED.timestamp`,
		blockHeight, blockHash, parentHash, stateRoot, blockSize, txCount, totalFee, avgTxSize, len(uniqueParticipants), timestamp)
	if err != nil {
		log.Printf("Error saving block to database: %v\n", err)
		return err
	}

	return nil
}

func getKeysFromMap(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
