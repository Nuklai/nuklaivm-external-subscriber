package grpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
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
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*emptypb.Empty, error) {
	fmt.Println("Initializing External Subscriber with genesis data...")

	// Decode genesis data
	genesisData := req.GetGenesis()
	var parsedGenesis map[string]interface{}
	if err := json.Unmarshal(genesisData, &parsedGenesis); err != nil {
		fmt.Println("Error parsing genesis data:", err)
	} else {
		fmt.Printf("Genesis Data (parsed): %v\n", parsedGenesis)
	}

	// Save genesis data to database
	_, err := s.db.Exec(`INSERT INTO genesis_data (data) VALUES ($1::json) ON CONFLICT DO NOTHING`, string(genesisData))
	if err != nil {
		fmt.Printf("Error saving genesis data to database: %v\n", err)
	}

	// Create parser from genesis bytes
	parser, err := vm.CreateParser(genesisData)
	if err != nil {
		fmt.Println("Error creating parser:", err)
		return nil, err
	}
	s.parser = parser

	return &emptypb.Empty{}, nil
}

// AcceptBlock processes a new block, saves relevant data to the database, and stores transactions and actions
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
	unitPrices := executedBlock.UnitPrices.String()
	timestamp := time.UnixMilli(blk.Tmstmp).Format(time.RFC3339)

	mu.Lock()
	defer mu.Unlock()

	// Save block data to the database
	_, err = s.db.Exec(`INSERT INTO blocks (block_height, block_hash, parent_block_hash, state_root, timestamp, unit_prices)
		VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (block_height) DO NOTHING`,
		blk.Hght, blockID, parentID, stateRoot, timestamp, unitPrices)
	if err != nil {
		fmt.Printf("Error saving block to database: %v\n", err)
		return &emptypb.Empty{}, nil
	}

	fmt.Printf("Block Details: ID: %s, ParentID: %s, StateRoot: %s, Height: %d\n", blockID, parentID, stateRoot, blk.Hght)
	fmt.Printf("Transactions in block %d: %d\n", blk.Hght, len(blk.Txs))

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

		fmt.Printf("\tTransaction %d: %s\n", i+1, txID)
		fmt.Printf("\tOutputs: %v\n", outputs)

		// Save transaction to the database
		_, err := s.db.Exec(`INSERT INTO transactions (tx_hash, block_hash, sponsor, max_fee, success, fee, outputs, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7::json, $8) ON CONFLICT (tx_hash) DO NOTHING`,
			txID, blockID, sponsor, tx.MaxFee(), success, fee, outputs, timestamp)
		if err != nil {
			fmt.Printf("Error saving transaction to database: %v\n", err)
			continue
		}

		fmt.Printf("\tNumber of Actions: %d\n", len(tx.Actions))

		// Process and save actions associated with the transaction
		for j, action := range tx.Actions {
			actionType := action.GetTypeID()

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

			fmt.Printf("\t\tAction %d: Type: %d, Details: %s\n", j+1, actionType, actionDetailsJSON)

			// Save each action to the actions table
			_, err = s.db.Exec(`INSERT INTO actions (tx_hash, action_type, action_details, timestamp)
				VALUES ($1, $2, $3::json, $4) ON CONFLICT (tx_hash) DO NOTHING`,
				txID, actionType, actionDetailsJSON, timestamp)
			if err != nil {
				fmt.Printf("Error saving action to database: %v\n", err)
			}
		}
	}

	return &emptypb.Empty{}, nil
}
