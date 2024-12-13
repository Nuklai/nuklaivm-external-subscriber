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
	"github.com/lib/pq"
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
	timestamp := time.UnixMilli(blk.Tmstmp).UTC().Format(time.RFC3339)
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
			DROP TABLE IF EXISTS blocks, transactions, actions, assets CASCADE;
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
		_, err = s.db.Exec(`
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

func getKeysFromMap(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
