package grpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
	"github.com/nuklai/nuklaivm/vm"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var mu sync.Mutex

// Server implements the ExternalSubscriberServer
type Server struct {
	pb.UnimplementedExternalSubscriberServer
	db     *sql.DB
	parser chain.Parser
}

// startGRPCServer starts the gRPC server for receiving block data
func StartGRPCServer(db *sql.DB, port string) {
	// Ensure the port has a colon prefix
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterExternalSubscriberServer(grpcServer, &Server{db: db})

	fmt.Printf("External Subscriber server is listening on port %s...\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Initialize receives genesis data for initialization and saves it to the database
// Initialize receives genesis data for initialization and saves it to the database
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*emptypb.Empty, error) {
	fmt.Println("Initializing External Subscriber with genesis data...")

	// Decode genesis data
	genesisData := req.GetGenesis()
	var parsedGenesis map[string]interface{}
	if err := json.Unmarshal(genesisData, &parsedGenesis); err != nil {
		fmt.Println("Error parsing genesis data:", err)
		return nil, err
	}

	mu.Lock()
	defer mu.Unlock()

	// Remove existing genesis data
	_, err := s.db.Exec(`DELETE FROM genesis_data`)
	if err != nil {
		fmt.Printf("Error deleting old genesis data from database: %v\n", err)
	}

	// Save new genesis data to the database
	_, err = s.db.Exec(`INSERT INTO genesis_data (data) VALUES ($1::json)`, string(genesisData))
	if err != nil {
		fmt.Printf("Error saving new genesis data to database: %v\n", err)
		return nil, err
	}

	// Create parser from genesis bytes
	parser, err := vm.CreateParser(genesisData)
	if err != nil {
		fmt.Println("Error creating parser:", err)
		return nil, err
	}
	s.parser = parser

	fmt.Println("Genesis data initialized successfully.")
	return &emptypb.Empty{}, nil
}

// AcceptBlock processes a new block, saves relevant data to the database, and stores transactions and actions
// AcceptBlock processes a new block, removes all data from subsequent blocks, and saves the new block's data
func (s *Server) AcceptBlock(ctx context.Context, req *pb.BlockRequest) (*emptypb.Empty, error) {
	fmt.Println("Received a new block:")

	// Extract and print raw BlockData
	blockData := req.GetBlockData()

	// Attempt to unmarshal the executed block using UnmarshalExecutedBlock
	executedBlock, err := chain.UnmarshalExecutedBlock(blockData, s.parser)
	if err != nil {
		fmt.Printf("Error parsing block data: %v\n", err)
		return &emptypb.Empty{}, nil
	}

	blk := executedBlock.Block
	blockID := executedBlock.BlockID.String()
	parentID := blk.Prnt.String()
	stateRoot := blk.StateRoot.String()
	timestamp := time.UnixMilli(blk.Tmstmp).Format(time.RFC3339)
	blockHeight := blk.Hght
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

	// Remove all data for blocks with height >= the current block height
	_, err = s.db.Exec(`DELETE FROM actions WHERE tx_hash IN (
		SELECT tx_hash FROM transactions WHERE block_hash IN (
			SELECT block_hash FROM blocks WHERE block_height >= $1
		)
	)`, blockHeight)
	if err != nil {
		fmt.Printf("Error deleting actions for blocks: %v\n", err)
	}

	_, err = s.db.Exec(`DELETE FROM transactions WHERE block_hash IN (
		SELECT block_hash FROM blocks WHERE block_height >= $1
	)`, blockHeight)
	if err != nil {
		fmt.Printf("Error deleting transactions for blocks: %v\n", err)
	}

	_, err = s.db.Exec(`DELETE FROM blocks WHERE block_height >= $1`, blockHeight)
	if err != nil {
		fmt.Printf("Error deleting blocks: %v\n", err)
	}

	fmt.Printf("Block Details: ID: %s, ParentID: %s, StateRoot: %s, Height: %d\n", blockID, parentID, stateRoot, blockHeight)
	fmt.Printf("Transactions in block %d: %d\n", blockHeight, len(blk.Txs))

	for i, tx := range blk.Txs {
		txID := tx.ID().String()
		sponsor := tx.Sponsor().String()
		fee := uint64(0)
		outputs := "{}"
		success := false

		if i < len(executedBlock.Results) {
			result := executedBlock.Results[i]
			fee = result.Fee
			success = result.Success

			if success && len(result.Outputs) > 0 {
				// Parse outputs if available
				packer := codec.NewReader(result.Outputs[0], len(result.Outputs[0]))
				r, err := vm.OutputParser.Unmarshal(packer)
				if err == nil {
					outputJSON, err := json.Marshal(r)
					if err == nil {
						outputs = string(outputJSON)
					}
				}
			}
		} else {
			fmt.Printf("Warning: No transaction result found for transaction index %d\n", i)
		}
		totalFee += fee
		uniqueParticipants[sponsor] = struct{}{}

		fmt.Printf("\tTransaction %d: %s\n", i+1, txID)
		fmt.Printf("\tOutputs: %v\n", outputs)

		// Save transaction to the database
		_, err := s.db.Exec(`INSERT INTO transactions (tx_hash, block_hash, sponsor, max_fee, success, fee, outputs, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7::json, $8)`,
			txID, blockID, sponsor, tx.MaxFee(), success, fee, outputs, timestamp)
		if err != nil {
			fmt.Printf("Error saving transaction to database: %v\n", err)
			continue
		}

		fmt.Printf("\tNumber of Actions: %d\n", len(tx.Actions))

		// Process and save actions associated with the transaction
		for j, action := range tx.Actions {
			actionType := action.GetTypeID()

			/*
				// Use reflection to dynamically get field names and values from the action
				actionValue := reflect.ValueOf(action).Elem()
				actionDetails := make(map[string]interface{})

				for i := 0; i < actionValue.NumField(); i++ {
					field := actionValue.Type().Field(i)
					fieldName := field.Name
					fieldValue := actionValue.Field(i).Interface()
					actionDetails[fieldName] = fieldValue
				}

				actionDetailsJSON, err := json.Marshal(actionDetails)
				if err != nil {
					fmt.Printf("Error marshaling action details: %v\n", err)
					actionDetailsJSON = []byte("{}")
				}
			*/
			actionDetailsJSON := "{}"
			if actionDetails, err := json.Marshal(action); err == nil {
				actionDetailsJSON = string(actionDetails)
			}

			fmt.Printf("\t\tAction %d: Type: %d, Details: %s\n", j+1, actionType, actionDetailsJSON)

			// Save each action to the actions table
			_, err = s.db.Exec(`INSERT INTO actions (tx_hash, action_type, action_details, timestamp)
				VALUES ($1, $2, $3::json, $4)`,
				txID, actionType, actionDetailsJSON, timestamp)
			if err != nil {
				fmt.Printf("Error saving action to database: %v\n", err)
			}
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
		blockHeight, blockID, parentID, stateRoot, blockSize, txCount, totalFee, avgTxSize, len(uniqueParticipants), timestamp)
	if err != nil {
		fmt.Printf("Error saving block to database: %v\n", err)
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, nil
}
