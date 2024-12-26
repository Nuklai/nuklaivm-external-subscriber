package server

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/ava-labs/hypersdk/chain"
	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
	"github.com/nuklai/nuklaivm/vm"
)

// handleInitialize processes the initialization request
func handleInitialize(dbConn *sql.DB, req *pb.InitializeRequest) (chain.Parser, error) {
	log.Println("Initializing External Subscriber with genesis data...")
	genesisData := req.GetGenesis()

	var parsedGenesis map[string]interface{}
	if err := json.Unmarshal(genesisData, &parsedGenesis); err != nil {
		log.Println("Error parsing genesis data:", err)
		return nil, err
	}

	_, err := dbConn.Exec(`DELETE FROM genesis_data`)
	if err != nil {
		log.Printf("Error deleting old genesis data from database: %v\n", err)
	}

	_, err = dbConn.Exec(`INSERT INTO genesis_data (data) VALUES ($1::json)`, string(genesisData))
	if err != nil {
		log.Printf("Error saving new genesis data to database: %v\n", err)
		return nil, err
	}

	parser, err := vm.CreateParser(genesisData)
	if err != nil {
		log.Println("Error creating parser:", err)
		return nil, err
	}

	log.Println("Genesis data initialized successfully.")
	return parser, nil
}
