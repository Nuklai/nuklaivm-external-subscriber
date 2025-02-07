// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/nuklai/nuklaivm-external-subscriber/api"
	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"github.com/nuklai/nuklaivm-external-subscriber/db"
	"github.com/nuklai/nuklaivm-external-subscriber/models"
	"github.com/nuklai/nuklaivm-external-subscriber/server"
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
	grpcPort := "50051"
	go server.StartGRPCServerWithRetries(database, grpcPort, 60)

	// Init the health monitor
	healthMonitor := api.InitHealthMonitor(database, grpcPort)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.SetTrustedProxies(nil)

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	// Health endpoint
	r.GET("/health", api.GetHealth(healthMonitor))                // Get the current health status
	r.GET("/health/history", api.GetHealthHistory(database))      // Get health insidents
	r.GET("/health/history/90days", api.Get90DayHealth(database)) // Get 90-day health history

	// Other endpoints
	r.GET("/genesis", api.GetGenesisData(database))

	r.GET("/blocks", api.GetAllBlocks(database))

	r.GET("/transactions", api.GetAllTransactions(database))
	r.GET("/transactions/:tx_hash", api.GetTransactionByHash(database))            // Fetch by transaction hash
	r.GET("/transactions/block/:identifier", api.GetTransactionsByBlock(database)) // Fetch transactions by block height or hash
	r.GET("/transactions/user/:user", api.GetTransactionsByUser(database))
	r.GET("/transactions/volumes", api.GetAllActionVolumes(database))
	r.GET("/transactions/volumes/:action_name", api.GetActionVolumesByName(database))
	r.GET("/transactions/volumes/total", api.GetTotalTransferVolume(database))                               // Fetch transactions by user with pagination
	r.GET("/transactions/estimated_fee/action_type/:action_type", api.GetEstimatedFeeByActionType(database)) // Fetch estimated fee by action type
	r.GET("/transactions/estimated_fee/action_name/:action_name", api.GetEstimatedFeeByActionName(database)) // Fetch estimated fee by action name
	r.GET("/transactions/estimated_fee", api.GetAggregateEstimatedFees(database))                            // Fetch aggregate estimated fees

	r.GET("/actions", api.GetAllActions(database))
	r.GET("/actions/:tx_hash", api.GetActionsByTransactionHash(database))     // Fetch actions by transaction hash
	r.GET("/actions/block/:identifier", api.GetActionsByBlock(database))      // Fetch actions by block height or hash
	r.GET("/actions/type/:action_type", api.GetActionsByActionType(database)) // Fetch actions by action type with pagination
	r.GET("/actions/name/:action_name", api.GetActionsByActionName(database)) // Fetch actions by action name with pagination
	r.GET("/actions/user/:user", api.GetActionsByUser(database))              // Fetch actions by user with pagination

	r.GET("/assets", api.GetAllAssets(database))
	r.GET("/assets/:asset_address", api.GetAssetByAddress(database))
	r.GET("/assets/type/:type", api.GetAssetsByType(database)) // Fetch assets by type
	r.GET("/assets/user/:user", api.GetAssetsByUser(database)) // Fetch assets by user

	r.GET("/validator_stake", api.GetAllValidatorStakes(database))
	r.GET("/validator_stake/:node_id", api.GetValidatorStakeByNodeID(database))

	// Start the health monitor (6s)
	go func() {
		ticker := time.NewTicker(6 * time.Second)
		defer ticker.Stop()

		var lastState models.HealthState
		var lastDate time.Time

		status := healthMonitor.GetHealthStatus()
		lastState = status.State
		lastDate = time.Now().UTC().Truncate(24 * time.Hour)

		for range ticker.C {
			status := healthMonitor.GetHealthStatus()
			currentDate := time.Now().UTC().Truncate(24 * time.Hour)

			if status.State != lastState {
				lastState = status.State
			}

			if !currentDate.Equal(lastDate) {
				lastDate = currentDate
				lastState = status.State
			}
		}
	}()

	r.GET("/accounts", api.GetAllAccounts(database))
	r.GET("/accounts/:address", api.GetAccountDetails(database))
	r.GET("/accounts/stats", api.GetAccountStats(database))

	// Start HTTP server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
