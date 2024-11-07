// Additional API endpoints to retrieve blockchain-related information from the database
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

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

// Define types for block, transaction, and action
type Block struct {
	ID              int    `json:"ID"`
	BlockHeight     int64  `json:"BlockHeight"`
	BlockHash       string `json:"BlockHash"`
	ParentBlockHash string `json:"ParentBlock"`
	StateRoot       string `json:"StateRoot"`
	Timestamp       string `json:"Timestamp"`
	UnitPrices      string `json:"UnitPrices"`
}

type Transaction struct {
	ID        int                    `json:"ID"`
	TxHash    string                 `json:"TxHash"`
	BlockHash string                 `json:"BlockHash"`
	Sponsor   string                 `json:"Sponsor"`
	MaxFee    float64                `json:"MaxFee"`
	Success   bool                   `json:"Success"`
	Fee       uint64                 `json:"Fee"`
	Outputs   map[string]interface{} `json:"Outputs"`
	Timestamp string                 `json:"Timestamp"`
}

type Action struct {
	ID            int                    `json:"ID"`
	TxHash        string                 `json:"TxHash"`
	ActionType    int                    `json:"ActionType"`
	ActionDetails map[string]interface{} `json:"ActionDetails"`
	Timestamp     string                 `json:"Timestamp"`
}

// GetGenesisData retrieves the genesis data stored in the database
func GetGenesisData(c *gin.Context) {
	var genesisData string
	err := db.QueryRow(`SELECT data FROM genesis_data LIMIT 1`).Scan(&genesisData)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Genesis data not found"})
		return
	}

	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(genesisData), &parsedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse genesis data"})
		return
	}

	c.JSON(http.StatusOK, parsedData)
}

// GetAllBlocks retrieves all blocks stored in the database, with pagination support
func GetAllBlocks(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	rows, err := db.Query(`SELECT * FROM blocks ORDER BY block_height DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve blocks"})
		return
	}
	defer rows.Close()

	var blocks []Block

	for rows.Next() {
		var block Block
		if err := rows.Scan(&block.ID, &block.BlockHeight, &block.BlockHash, &block.ParentBlockHash, &block.StateRoot, &block.Timestamp, &block.UnitPrices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse block"})
			return
		}
		blocks = append(blocks, block)
	}

	c.JSON(http.StatusOK, blocks)
}

// GetAllTransactions retrieves all transactions stored in the database, with pagination support
func GetAllTransactions(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	rows, err := db.Query(`SELECT * FROM transactions ORDER BY timestamp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions"})
		return
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var tx Transaction
		var outputsJSON []byte
		if err := rows.Scan(&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.MaxFee, &tx.Success, &tx.Fee, &outputsJSON, &tx.Timestamp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse transaction"})
			return
		}
		if err := json.Unmarshal(outputsJSON, &tx.Outputs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse outputs"})
			return
		}
		transactions = append(transactions, tx)
	}

	c.JSON(http.StatusOK, transactions)
}

// GetAllActions retrieves all actions stored in the database, with pagination support
func GetAllActions(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	rows, err := db.Query(`SELECT * FROM actions ORDER BY timestamp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve actions"})
		return
	}
	defer rows.Close()

	var actions []Action

	for rows.Next() {
		var action Action
		var actionDetailsJSON []byte
		if err := rows.Scan(&action.ID, &action.TxHash, &action.ActionType, &actionDetailsJSON, &action.Timestamp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse action"})
			return
		}
		if err := json.Unmarshal(actionDetailsJSON, &action.ActionDetails); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse action details"})
			return
		}
		actions = append(actions, action)
	}

	c.JSON(http.StatusOK, actions)
}

func GetBlock(c *gin.Context) {
	blockIdentifier := c.Param("identifier")

	// Try to parse blockIdentifier as an integer for block height lookup
	var query string
	var err error
	var block Block

	// Check if blockIdentifier is numeric to determine if it's a block height
	if height, parseErr := strconv.ParseInt(blockIdentifier, 10, 64); parseErr == nil {
		// blockIdentifier is a number; use it to query by block height
		query = `
			SELECT id, block_height, block_hash, parent_block_hash, state_root, timestamp, unit_prices
			FROM blocks
			WHERE block_height = $1::bigint`
		err = db.QueryRow(query, height).Scan(
			&block.ID, &block.BlockHeight, &block.BlockHash, &block.ParentBlockHash, &block.StateRoot, &block.Timestamp, &block.UnitPrices)
	} else {
		// blockIdentifier is not a number; use it to query by block hash
		query = `
			SELECT id, block_height, block_hash, parent_block_hash, state_root, timestamp, unit_prices
			FROM blocks
			WHERE block_hash = $1`
		err = db.QueryRow(query, blockIdentifier).Scan(
			&block.ID, &block.BlockHeight, &block.BlockHash, &block.ParentBlockHash, &block.StateRoot, &block.Timestamp, &block.UnitPrices)
	}

	// Check for query errors
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Block not found"})
		return
	}

	c.JSON(http.StatusOK, block)
}

// GetTransactionByHash retrieves a transaction by its hash
func GetTransactionByHash(c *gin.Context) {
	txHash := c.Param("tx_hash")
	var tx Transaction
	var outputsJSON []byte
	err := db.QueryRow(`SELECT * FROM transactions WHERE tx_hash = $1`, txHash).Scan(
		&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.MaxFee, &tx.Success, &tx.Fee, &outputsJSON, &tx.Timestamp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	if err := json.Unmarshal(outputsJSON, &tx.Outputs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse outputs"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// GetTransactionsByBlock retrieves transactions by block number or hash
func GetTransactionsByBlock(c *gin.Context) {
	blockIdentifier := c.Param("identifier")

	var query string
	var rows *sql.Rows
	var err error

	// Check if blockIdentifier is numeric for querying by block height
	if height, parseErr := strconv.ParseInt(blockIdentifier, 10, 64); parseErr == nil {
		// Query by block height using a join with the blocks table
		query = `
			SELECT transactions.id, transactions.tx_hash, transactions.block_hash, transactions.sponsor, transactions.max_fee, transactions.success, transactions.fee, transactions.outputs, transactions.timestamp
			FROM transactions
			INNER JOIN blocks ON transactions.block_hash = blocks.block_hash
			WHERE blocks.block_height = $1`
		rows, err = db.Query(query, height)
	} else {
		// Query by block hash directly
		query = `
			SELECT id, tx_hash, block_hash, sponsor, max_fee, success, fee, outputs, timestamp
			FROM transactions
			WHERE block_hash = $1`
		rows, err = db.Query(query, blockIdentifier)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve transactions"})
		return
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var tx Transaction
		var outputsJSON []byte
		if err := rows.Scan(&tx.ID, &tx.TxHash, &tx.BlockHash, &tx.Sponsor, &tx.MaxFee, &tx.Success, &tx.Fee, &outputsJSON, &tx.Timestamp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse transaction"})
			return
		}
		if err := json.Unmarshal(outputsJSON, &tx.Outputs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse outputs"})
			return
		}
		transactions = append(transactions, tx)
	}

	c.JSON(http.StatusOK, transactions)
}

// GetActionsByBlock retrieves actions by block number or hash
func GetActionsByBlock(c *gin.Context) {
	blockIdentifier := c.Param("identifier")

	var query string
	var rows *sql.Rows
	var err error

	// Check if blockIdentifier is numeric for querying by block height
	if height, parseErr := strconv.ParseInt(blockIdentifier, 10, 64); parseErr == nil {
		// Query by block height using a join with the blocks table
		query = `
			SELECT actions.id, actions.tx_hash, actions.action_type, actions.action_details, actions.timestamp
			FROM actions
			INNER JOIN transactions ON actions.tx_hash = transactions.tx_hash
			INNER JOIN blocks ON transactions.block_hash = blocks.block_hash
			WHERE blocks.block_height = $1`
		rows, err = db.Query(query, height)
	} else {
		// Query by block hash directly
		query = `
			SELECT actions.id, actions.tx_hash, actions.action_type, actions.action_details, actions.timestamp
			FROM actions
			INNER JOIN transactions ON actions.tx_hash = transactions.tx_hash
			WHERE transactions.block_hash = $1`
		rows, err = db.Query(query, blockIdentifier)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve actions"})
		return
	}
	defer rows.Close()

	var actions []Action

	for rows.Next() {
		var action Action
		var actionDetailsJSON []byte
		if err := rows.Scan(&action.ID, &action.TxHash, &action.ActionType, &actionDetailsJSON, &action.Timestamp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse action"})
			return
		}
		if err := json.Unmarshal(actionDetailsJSON, &action.ActionDetails); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse action details"})
			return
		}
		actions = append(actions, action)
	}

	c.JSON(http.StatusOK, actions)
}

// GetActionsByTransactionHash retrieves actions by transaction hash
func GetActionsByTransactionHash(c *gin.Context) {
	txHash := c.Param("tx_hash")
	rows, err := db.Query(`SELECT id, tx_hash, action_type, action_details, timestamp FROM actions WHERE tx_hash = $1`, txHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve actions"})
		return
	}
	defer rows.Close()

	var actions []Action

	for rows.Next() {
		var action Action
		var actionDetailsJSON []byte
		if err := rows.Scan(&action.ID, &action.TxHash, &action.ActionType, &actionDetailsJSON, &action.Timestamp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse action"})
			return
		}
		if err := json.Unmarshal(actionDetailsJSON, &action.ActionDetails); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse action details"})
			return
		}
		actions = append(actions, action)
	}

	c.JSON(http.StatusOK, actions)
}

// startGRPCServer starts the gRPC server for receiving block data
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

// main function to register routes and start servers
func main() {
	var err error
	connStr := "postgres://postgres:postgres@localhost:5432/blockchain?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	go startGRPCServer()

	r := gin.Default()

	// Endpoint registration
	r.GET("/genesis", GetGenesisData)
	r.GET("/blocks", GetAllBlocks)
	r.GET("/blocks/:identifier", GetBlock) // Fetch by block height or hash
	r.GET("/transactions", GetAllTransactions)
	r.GET("/transactions/:tx_hash", GetTransactionByHash)
	r.GET("/transactions/block/:identifier", GetTransactionsByBlock) // Fetch transactions by block height or hash
	r.GET("/actions", GetAllActions)
	r.GET("/actions/:tx_hash", GetActionsByTransactionHash) // Fetch actions by transaction hash
	r.GET("/actions/block/:identifier", GetActionsByBlock)  // Fetch actions by block height or hash

	r.Run()
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

// AcceptBlock processes a new block when it's accepted and stores relevant data in the database
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
	_, err = db.Exec(`INSERT INTO blocks (block_height, block_hash, parent_block_hash, state_root, timestamp, unit_prices)
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
		_, err := db.Exec(`INSERT INTO transactions (tx_hash, block_hash, sponsor, max_fee, success, fee, outputs, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8) ON CONFLICT (tx_hash) DO NOTHING`,
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
			_, err = db.Exec(`INSERT INTO actions (tx_hash, action_type, action_details, timestamp)
				VALUES ($1, $2, $3::jsonb, $4) ON CONFLICT (tx_hash) DO NOTHING`,
				txID, actionType, actionDetailsJSON, timestamp)
			if err != nil {
				fmt.Printf("Error saving action to database: %v\n", err)
			}
		}
	}

	return &emptypb.Empty{}, nil
}
