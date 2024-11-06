package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
	"github.com/nuklai/nuklaivm/vm"
)

var db *sql.DB
var mu sync.Mutex

// Server implements the ExternalSubscriberServer
type Server struct {
	pb.UnimplementedExternalSubscriberServer
	parser chain.Parser
}

func main() {
	// Database connection setup
	var err error
	connStr := "postgres://postgres:postgres@localhost:5432/blockchain?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error connecting to the database: %v\n", err)
		os.Exit(1)
	}

	// Start gRPC server for receiving block data
	go startGRPCServer()

	// Start REST API server
	r := gin.Default()

	// API endpoints
	r.GET("/blocks/:height", GetBlockByHeight)
	r.GET("/transactions/:tx_hash", GetTransactionByHash)
	r.GET("/actions/:tx_id", GetActionsByTransactionID)

	// Start the REST API server
	r.Run() // Default is :8080
}

func startGRPCServer() {
	const port = ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterExternalSubscriberServer(grpcServer, &Server{})

	fmt.Printf("External Subscriber server is listening on port %s...\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Initialize receives genesis data for initialization
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
	_, err := db.Exec(`INSERT INTO genesis_data (data) VALUES ($1::jsonb) ON CONFLICT DO NOTHING`, string(genesisData))
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

// AcceptBlock prints block data whenever a block is accepted
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

	// Extract details from the unmarshaled block
	blk := executedBlock.Block
	blockID := executedBlock.BlockID.String()
	parentID := blk.Prnt.String()
	stateRoot := blk.StateRoot.String()
	unitPrices := executedBlock.UnitPrices.String()

	fmt.Printf("Block ID: %s\n", blockID)
	fmt.Printf("Parent ID: %s\n", parentID)
	fmt.Printf("Timestamp: %s\n", time.UnixMilli(blk.Tmstmp).Format(time.RFC3339))
	fmt.Printf("Height: %d\n", blk.Hght)
	fmt.Printf("State Root: %s\n", stateRoot)
	fmt.Printf("Unit Prices: %+v\n", unitPrices)

	// Mutex to prevent duplicate data insertion from multiple nodes
	mu.Lock()
	defer mu.Unlock()

	// Save block data to database
	_, err = db.Exec(`INSERT INTO blocks (block_height, block_hash, parent_block_hash, state_root, timestamp, unit_prices) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (block_hash) DO NOTHING`,
		blk.Hght, blockID, parentID, stateRoot, time.UnixMilli(blk.Tmstmp).Format(time.RFC3339), unitPrices)
	if err != nil {
		fmt.Printf("Error saving block to database: %v\n", err)
		return &emptypb.Empty{}, nil
	}

	// Save transaction data to database and process transaction results
	if len(blk.Txs) > 0 {
		fmt.Printf("Transactions in block %d:\n", blk.Hght)
		for i, tx := range blk.Txs {
			fmt.Printf("\tTransaction %d:\n", i)
			success := false
			txID := tx.ID().String()
			sponsor := tx.Sponsor().String()
			fee := uint64(0)
			outputs := "{}"

			// Process transaction results if available
			if len(executedBlock.Results) > i {
				result := executedBlock.Results[i]
				fee = result.Fee
				fmt.Printf("\tFee Consumed: %d NAI\n", fee)
				if result.Success {
					success = true
					packer := codec.NewReader(result.Outputs[0], len(result.Outputs[0]))
					r, err := vm.OutputParser.Unmarshal(packer)
					if err == nil {
						outputJSON, err := json.Marshal(r)
						if err == nil {
							outputs = string(outputJSON)
						}
					}
				}
			}

			fmt.Printf("\tOutputs: %v\n", outputs)

			// Save transaction to the database
			_, err := db.Exec(`INSERT INTO transactions (tx_hash, block_hash, sponsor, max_fee, success, fee, outputs, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8) ON CONFLICT (tx_hash) DO NOTHING`,
				txID, blockID, sponsor, tx.MaxFee(), success, fee, outputs, time.UnixMilli(blk.Tmstmp).Format(time.RFC3339))
			if err != nil {
				fmt.Printf("Error saving transaction to database: %v\n", err)
			}
		}
	}

	return &emptypb.Empty{}, nil
}

// GetBlockByHeight retrieves a block by its height
func GetBlockByHeight(c *gin.Context) {
	height := c.Param("height")
	var block struct {
		ID          int
		BlockHeight int64
		BlockHash   string
		ParentBlock string
		StateRoot   string
		Timestamp   string
		UnitPrices  string
	}
	err := db.QueryRow(`SELECT * FROM blocks WHERE block_height = $1`, height).Scan(
		&block.ID, &block.BlockHeight, &block.BlockHash, &block.ParentBlock, &block.StateRoot, &block.Timestamp, &block.UnitPrices)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
		return
	}

	c.JSON(http.StatusOK, block)
}

// GetTransactionByHash retrieves a transaction by its hash
func GetTransactionByHash(c *gin.Context) {
	txHash := c.Param("tx_hash")
	var tx struct {
		ID        int
		TxHash    string
		BlockHash string
		Sponsor   string
		MaxFee    float64
		Success   bool
		Fee       uint64
		Outputs   string
		Timestamp string
	}
	err := db.QueryRow(`SELECT * FROM transactions WHERE tx_hash = $1`, txHash).Scan(
		&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.MaxFee, &tx.Success, &tx.Fee, &tx.Outputs, &tx.Timestamp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// GetActionsByTransactionID retrieves actions by transaction ID
func GetActionsByTransactionID(c *gin.Context) {
	txID := c.Param("tx_id")
	rows, err := db.Query(`SELECT * FROM actions WHERE tx_hash = $1`, txID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve actions"})
		return
	}
	defer rows.Close()

	var actions []struct {
		ID            int
		TxID          int
		ActionType    int
		ActionDetails string
	}

	for rows.Next() {
		var action struct {
			ID            int
			TxID          int
			ActionType    int
			ActionDetails string
		}
		if err := rows.Scan(&action.ID, &action.TxID, &action.ActionType, &action.ActionDetails); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse action"})
			return
		}
		actions = append(actions, action)
	}

	c.JSON(http.StatusOK, actions)
}
