package config

import (
	"os"
)

// GetDatabaseURL retrieves the database connection string
func GetDatabaseURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/blockchain?sslmode=disable"
		// log.Fatal("DATABASE_URL environment variable not set")
	}
	return url
}
