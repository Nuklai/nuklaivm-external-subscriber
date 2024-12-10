// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
)

// GetDatabaseURL retrieves the database connection string
func GetDatabaseURL() string {
	// Retrieve env variable for DB_HOST and set default value if not present

	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	user := GetEnv("DB_USER", "postgres")
	password := GetEnv("DB_PASSWORD", "postgres")
	dbName := GetEnv("DB_NAME", "nuklaivm")
	sslMode := GetEnv("DB_SSL_MODE", "disable")

	// Encode password to handle special characters
	encodedPassword := url.QueryEscape(password)

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, encodedPassword, host, port, dbName, sslMode,
	)
}

// GetWhitelistIPs retrieves the list of whitelisted IPs from the environment variable
// and resolves domain names to IPs.
// GetWhitelistIPs retrieves the list of whitelisted IPs and CIDR ranges
func GetWhitelistIPs() ([]string, []string) {
	ipList := GetEnv("GRPC_WHITELISTED_BLOCKCHAIN_NODES", "127.0.0.1,localhost,::1")
	entries := strings.Split(ipList, ",")

	whitelistIPs := []string{}
	whitelistCIDRs := []string{}
	defaultEntries := []string{"127.0.0.1", "localhost", "::1"}

	// Combine default entries and user-provided entries
	for _, entry := range append(defaultEntries, entries...) {
		entry = strings.TrimSpace(entry)
		if strings.Contains(entry, "/") {
			// CIDR range
			whitelistCIDRs = append(whitelistCIDRs, entry)
		} else {
			// IP or domain
			ips, err := resolveToIPs(entry)
			if err == nil {
				whitelistIPs = append(whitelistIPs, ips...)
			} else {
				log.Printf("Failed to resolve: %s, skipping", entry)
			}
		}
	}

	return uniqueStrings(whitelistIPs), uniqueStrings(whitelistCIDRs)
}

// GetEnv retrieves the value of the environment variable named by the key.
// If the variable is not present, it returns the default value.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// resolveToIPs resolves a domain name to its IP addresses or directly returns the IP if it's already valid
func resolveToIPs(host string) ([]string, error) {
	// Check if the host is already a valid IP
	if net.ParseIP(host) != nil {
		return []string{host}, nil
	}

	// Attempt to resolve the domain name
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	// Convert net.IP to strings
	ipStrings := []string{}
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}
	return ipStrings, nil
}

// uniqueStrings removes duplicates from a slice of strings
func uniqueStrings(slice []string) []string {
	unique := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if _, exists := unique[item]; !exists {
			unique[item] = true
			result = append(result, item)
		}
	}
	return result
}
