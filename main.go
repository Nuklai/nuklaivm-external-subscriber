package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/vm"
)

// Server implements the ExternalSubscriberServer
type Server struct {
	pb.UnimplementedExternalSubscriberServer
	parser chain.Parser
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

	// Extract and print details from the unmarshaled block
	blk := executedBlock.Block
	fmt.Printf("Block ID: %s\n", executedBlock.BlockID)
	fmt.Printf("Parent ID: %s\n", blk.Prnt)
	fmt.Printf("Timestamp: %s\n", time.UnixMilli(blk.Tmstmp).Format(time.RFC3339))
	fmt.Printf("Height: %d\n", blk.Hght)
	fmt.Printf("State Root: %s\n", blk.StateRoot)
	fmt.Printf("Unit Prices: %+v\n", executedBlock.UnitPrices)

	// Print detailed transaction information
	if len(blk.Txs) > 0 {
		fmt.Printf("Transactions: %d\n", len(blk.Txs))
		for i, tx := range blk.Txs {
			fmt.Printf("  Transaction %d:\n", i+1)
			fmt.Printf("    ID: %s\n", tx.ID())
			fmt.Printf("    Sponsor: %s\n", tx.Sponsor())
			fmt.Printf("    Max Fee: %d\n", tx.MaxFee())

			// Display actions and their types
			if len(tx.Actions) > 0 {
				fmt.Printf("    Actions: %d\n", len(tx.Actions))
				for j, action := range tx.Actions {
					fmt.Printf("      Action %d Type ID: %d\n", j+1, action.GetTypeID())
					// Additional details for each action can be added here
				}
			} else {
				fmt.Println("    No actions in transaction.")
			}
		}
	} else {
		fmt.Println("No transactions in block.")
	}

	// Print detailed output
	if len(executedBlock.Results) > 0 {
		processResult(executedBlock.Results[0])
	}

	return &emptypb.Empty{}, nil
}

// Helper function to process the transaction result
func processResult(result *chain.Result) error {
	if result != nil && result.Success {
		utils.Outf("{{green}}fee consumed:{{/}} %s NAI\n", utils.FormatBalance(result.Fee))

		// Use NewReader to create a Packer from the result output
		packer := codec.NewReader(result.Outputs[0], len(result.Outputs[0]))
		r, err := vm.OutputParser.Unmarshal(packer)
		if err != nil {
			return err
		}

		// Assert the output to the expected type
		output, ok := r.(interface{})
		if !ok {
			return errors.New("failed to assert typed output to expected result type")
		}

		// Output the results
		utils.Outf("{{green}}output: {{/}} %+v\n", output)
	}
	return nil
}

func main() {
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
