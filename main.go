// Additional API endpoints to retrieve blockchain-related information from the database
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/nuklai/nuklaivm-external-subscriber/api"
	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"github.com/nuklai/nuklaivm-external-subscriber/db"
	"github.com/nuklai/nuklaivm-external-subscriber/grpc"
)

// main function to register routes and start servers
func main() {
	// Initialize the database
	connStr := config.GetDatabaseURL()
	database, err := db.InitDB(connStr)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Start the gRPC server
	go grpc.StartGRPCServer(database, "50051")

	// Setup Gin router
	r := gin.Default()

	// Health endpoint
	r.GET("/health", api.HealthCheck())

	// Other endpoints
	r.GET("/genesis", api.GetGenesisData(database))

	r.GET("/blocks", api.GetAllBlocks(database))
	r.GET("/blocks/:identifier", api.GetBlock(database)) // Fetch by block height or hash

	r.GET("/transactions", api.GetAllTransactions(database))
	r.GET("/transactions/:tx_hash", api.GetTransactionByHash(database))            // Fetch by transaction hash
	r.GET("/transactions/block/:identifier", api.GetTransactionsByBlock(database)) // Fetch transactions by block height or hash
	r.GET("/transactions/user/:user", api.GetTransactionsByUser(database))         // Fetch transactions by user with pagination

	r.GET("/actions", api.GetAllActions(database))
	r.GET("/actions/:tx_hash", api.GetActionsByTransactionHash(database)) // Fetch actions by transaction hash
	r.GET("/actions/block/:identifier", api.GetActionsByBlock(database))  // Fetch actions by block height or hash
	r.GET("/actions/user/:user", api.GetActionsByUser(database))          // Fetch actions by user with pagination

	r.GET("/assets", api.GetAllAssets(database))
	r.GET("/assets/type/:type", api.GetAssetsByType(database)) // Fetch assets by type
	r.GET("/assets/user/:user", api.GetAssetsByUser(database)) // Fetch assets by user

	// Start HTTP server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
