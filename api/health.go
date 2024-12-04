// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// checkGRPC checks the reachability of the gRPC server.
func checkGRPC(grpcPort string) string {
	conn, err := net.DialTimeout("tcp", grpcPort, 2*time.Second)
	if err != nil {
		log.Printf("gRPC connection failed: %v", err)
		return "unreachable"
	}
	defer conn.Close()
	return "reachable"
}

// checkDatabase checks the reachability of the database.
func checkDatabase(db *sql.DB) string {
	if err := db.Ping(); err != nil {
		log.Printf("Database connection failed: %v", err)
		return "unreachable"
	}
	return "reachable"
}

// HealthCheck performs a comprehensive health check of the subscriber.
func HealthCheck(grpcPort string, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		grpcStatus := checkGRPC(grpcPort)
		databaseStatus := checkDatabase(db)

		// Determine the overall status
		status := "ok"
		if grpcStatus == "unreachable" || databaseStatus == "unreachable" {
			status = "service unavailable"
		}

		// Return the health status
		httpStatusCode := http.StatusOK
		if status == "service unavailable" {
			httpStatusCode = http.StatusServiceUnavailable
		}

		c.JSON(httpStatusCode, gin.H{
			"status": status,
			"details": gin.H{
				"database": databaseStatus,
				"grpc":     grpcStatus,
			},
		})
	}
}
