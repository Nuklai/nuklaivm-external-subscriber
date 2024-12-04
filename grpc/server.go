// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package grpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
	"github.com/nuklai/nuklaivm-external-subscriber/consts"
	subscriberDB "github.com/nuklai/nuklaivm-external-subscriber/db"
	"github.com/nuklai/nuklaivm/vm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var mu sync.Mutex

// Server implements the ExternalSubscriberServer
type Server struct {
	pb.UnimplementedExternalSubscriberServer
	db     *sql.DB
	parser chain.Parser
}

// StartGRPCServer starts the gRPC server for receiving block data
func StartGRPCServer(db *sql.DB, port string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in StartGRPCServer: %v", r)
		}
	}()

	// Load the whitelist
	LoadWhitelist()

	// Ensure the port has a colon prefix
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Failed to listen on port %s: %v", port, err)
		return err
	}

	// Use insecure credentials to allow plaintext communication
	serverOptions := []grpc.ServerOption{
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(UnaryInterceptor), // Attach the interceptor
	}
	grpcServer := grpc.NewServer(serverOptions...)

	// Register your ExternalSubscriber service
	pb.RegisterExternalSubscriberServer(grpcServer, &Server{db: db})

	// Enable gRPC reflection for tools like grpcurl
	reflection.Register(grpcServer)

	log.Printf("External Subscriber server is listening on port %s...\n", port)
	return grpcServer.Serve(lis)
}

// StartGRPCServerWithRetries retries gRPC server startup in case of failure
func StartGRPCServerWithRetries(db *sql.DB, port string, retries int) {
	for i := 0; i < retries; i++ {
		err := StartGRPCServer(db, port)
		if err != nil {
			log.Printf("gRPC server failed to start: %v. Retrying (%d/%d)...", err, i+1, retries)
			time.Sleep(5 * time.Second)
			continue
		}
		return // Successful start
	}
	log.Fatal("gRPC server failed to start after maximum retries")
}

// Initialize receives genesis data for initialization and saves it to the database
// Initialize receives genesis data for initialization and saves it to the database
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in Initialize: %v", r)
		}
	}()

	log.Println("Initializing External Subscriber with genesis data...")

	// Decode genesis data
	genesisData := req.GetGenesis()
	var parsedGenesis map[string]interface{}
	if err := json.Unmarshal(genesisData, &parsedGenesis); err != nil {
		log.Println("Error parsing genesis data:", err)
		return nil, err
	}

	mu.Lock()
	defer mu.Unlock()

	// Remove existing genesis data
	_, err := s.db.Exec(`DELETE FROM genesis_data`)
	if err != nil {
		log.Printf("Error deleting old genesis data from database: %v\n", err)
	}

	// Save new genesis data to the database
	_, err = s.db.Exec(`INSERT INTO genesis_data (data) VALUES ($1::json)`, string(genesisData))
	if err != nil {
		log.Printf("Error saving new genesis data to database: %v\n", err)
		return nil, err
	}

	// Create parser from genesis bytes
	parser, err := vm.CreateParser(genesisData)
	if err != nil {
		log.Println("Error creating parser:", err)
		return nil, err
	}
	s.parser = parser

	log.Println("Genesis data initialized successfully.")
	return &emptypb.Empty{}, nil
}

// AcceptBlock processes a new block, saves relevant data to the database, and stores transactions and actions
func (s *Server) AcceptBlock(ctx context.Context, req *pb.BlockRequest) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in AcceptBlock: %v", r)
		}
	}()

	if s.parser == nil {
		log.Println("Parser is not initialized. Rejecting the request.")
		return nil, errors.New("parser not initialized")
	}

	blockData := req.GetBlockData()

	// Attempt to unmarshal the executed block using UnmarshalExecutedBlock
	executedBlock, err := chain.UnmarshalExecutedBlock(blockData, s.parser)
	if err != nil {
		log.Printf("Error parsing block data: %v\n", err)
		return &emptypb.Empty{}, nil
	}

	blk := executedBlock.Block
	blockHeight := blk.Hght
	blockHash := executedBlock.BlockID.String()
	parentHash := blk.Prnt.String()
	stateRoot := blk.StateRoot.String()
	timestamp := time.UnixMilli(blk.Tmstmp).Format(time.RFC3339)
	blockSize := blk.Size()
	txCount := len(blk.Txs)
	avgTxSize := 0.0
	if txCount > 0 {
		avgTxSize = float64(blockSize) / float64(txCount)
	}

	uniqueParticipants := make(map[string]struct{})
	totalFee := uint64(0)

	mu.Lock()
	defer mu.Unlock()

	// If the block height is 1, reset the database schema
	if blockHeight == 1 {
		log.Println("First block detected (genesis). Resetting the database...")

		// Drop all tables
		_, err := s.db.Exec(`
			DROP TABLE IF EXISTS actions, transactions, blocks CASCADE;
		`)
		if err != nil {
			log.Printf("Error dropping existing tables: %v\n", err)
			return &emptypb.Empty{}, nil
		}

		// Re-create the schema
		err = subscriberDB.CreateSchema(s.db)
		if err != nil {
			log.Printf("Error re-creating schema: %v\n", err)
			return &emptypb.Empty{}, nil
		}

		log.Println("Database reset and schema re-created successfully.")
	}

	log.Printf("Block Details: Height: %d, Hash: %s, ParentHash: %s, Transactions: %d\n", blockHeight, blockHash, parentHash, len(blk.Txs))

	for i, tx := range blk.Txs {
		txID := tx.ID().String()
		sponsor := tx.Sponsor().String()
		fee := uint64(0)
		outputs := []map[string]interface{}{}
		success := false

		actions := []map[string]interface{}{}

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
			actionName := consts.ActionNames[actionType]

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
			_, err := s.db.Exec(`
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

			if actionType == 4 { // create_asset action
				// Parse actionInputJSON into map[string]interface{}
				var actionInput map[string]interface{}
				err := json.Unmarshal([]byte(actionInputJSON), &actionInput)
				if err != nil {
					log.Printf("Error unmarshaling action input: %v\n", err)
					continue
				}
				actionOutput := outputs[j]
				assetID := actionOutput["asset_id"].(string)
				assetTypeID := actionInput["asset_type"].(float64)
				assetType := map[float64]string{0: "fungible", 1: "non-fungible", 2: "fractional"}[assetTypeID]

				// Insert asset into the assets table
				_, err = s.db.Exec(`
        INSERT INTO assets (
            asset_id, asset_type_id, asset_type, asset_creator, tx_hash, name, symbol, decimals, metadata, max_supply, mint_admin, pause_unpause_admin, freeze_unfreeze_admin, enable_disable_kyc_account_admin, timestamp
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        ON CONFLICT (asset_id) DO UPDATE
        SET asset_type_id = EXCLUDED.asset_type_id,
            asset_type = EXCLUDED.asset_type,
            asset_creator = EXCLUDED.asset_creator,
            tx_hash = EXCLUDED.tx_hash,
            name = EXCLUDED.name,
            symbol = EXCLUDED.symbol,
            decimals = EXCLUDED.decimals,
            metadata = EXCLUDED.metadata,
            max_supply = EXCLUDED.max_supply,
            mint_admin = EXCLUDED.mint_admin,
            pause_unpause_admin = EXCLUDED.pause_unpause_admin,
            freeze_unfreeze_admin = EXCLUDED.freeze_unfreeze_admin,
            enable_disable_kyc_account_admin = EXCLUDED.enable_disable_kyc_account_admin,
            timestamp = EXCLUDED.timestamp
    `, assetID, assetTypeID, assetType, sponsor, txID,
					actionInput["name"], actionInput["symbol"], actionInput["decimals"],
					actionInput["metadata"], actionInput["max_supply"], actionInput["mint_admin"],
					actionInput["pause_unpause_admin"], actionInput["freeze_unfreeze_admin"],
					actionInput["enable_disable_kyc_account_admin"], timestamp)
				if err != nil {
					log.Printf("Error saving asset to database: %v\n", err)
				}
			}
		}

		// Convert actions to JSON for storing in the transactions table
		actionsJSON, err := json.Marshal(actions)
		if err != nil {
			log.Printf("Error marshaling actions: %v\n", err)
			continue
		}

		// Save the transaction with aggregated actions
		_, err = s.db.Exec(`
				INSERT INTO transactions (tx_hash, block_hash, sponsor, max_fee, success, fee, actions, timestamp)
				VALUES ($1, $2, $3, $4, $5, $6, $7::json, $8)
				ON CONFLICT (tx_hash) DO UPDATE
				SET block_hash = EXCLUDED.block_hash,
						sponsor = EXCLUDED.sponsor,
						max_fee = EXCLUDED.max_fee,
						success = EXCLUDED.success,
						fee = EXCLUDED.fee,
						actions = EXCLUDED.actions,
						timestamp = EXCLUDED.timestamp`,
			txID, blockHash, sponsor, tx.MaxFee(), success, fee, actionsJSON, timestamp)
		if err != nil {
			log.Printf("Error saving transaction to database: %v\n", err)
		}
	}

	// Save the new block data to the database
	_, err = s.db.Exec(`
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
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, nil
}
