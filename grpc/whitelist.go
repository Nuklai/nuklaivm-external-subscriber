// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	WhitelistedIPs   = make(map[string]bool)
	WhitelistedCIDRs []*net.IPNet
)

// LoadWhitelist loads the whitelist using the config package
func LoadWhitelist() {
	ips, cidrs := config.GetWhitelistIPs()

	// Load individual IPs
	for _, ip := range ips {
		WhitelistedIPs[ip] = true
	}

	// Load CIDR ranges
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("Invalid CIDR range: %s", cidr)
			continue
		}
		WhitelistedCIDRs = append(WhitelistedCIDRs, ipNet)
	}

	log.Printf("Loaded whitelisted IPs: %v", WhitelistedIPs)
	log.Printf("Loaded whitelisted CIDRs: %v", cidrs)
}

// isAllowedIP checks if an IP is whitelisted
func isAllowedIP(clientIP string) bool {
	// Check against individual IPs
	if WhitelistedIPs[clientIP] {
		return true
	}

	// Check against CIDR ranges
	ip := net.ParseIP(clientIP)
	for _, ipNet := range WhitelistedCIDRs {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

// UnaryInterceptor checks the IP of the client and allows/denies the connection
func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve peer info")
	}

	clientIP := strings.Split(peerInfo.Addr.String(), ":")[0]
	if !isAllowedIP(clientIP) {
		log.Printf("Unauthorized connection attempt from IP: %s", clientIP)
		return nil, fmt.Errorf("unauthorized IP: %s", clientIP)
	}

	return handler(ctx, req)
}
