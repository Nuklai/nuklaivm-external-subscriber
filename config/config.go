package config

import (
	"fmt"
	"os"
)

// GetDatabaseURL retrieves the database connection string
func GetDatabaseURL() string {
	// Retrieve env variable for DB_HOST and set default value if not present

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "nuklaivm")
	sslMode := getEnv("DB_SSL_MODE", "disable")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode,
	)
}

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present, it returns the default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
