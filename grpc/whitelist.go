package grpc

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var WhitelistedIPs = make(map[string]bool)

// LoadWhitelist loads the whitelist using the config package
func LoadWhitelist() {
	ips := config.GetWhitelistIPs()
	if len(ips) == 0 {
		log.Println("No whitelisted IPs provided. The gRPC server will reject all connections.")
		return
	}

	// Populate the whitelist map
	for _, ip := range ips {
		WhitelistedIPs[ip] = true
	}

	log.Printf("Loaded whitelisted IPs: %v\n", WhitelistedIPs)
}

// UnaryInterceptor checks the IP of the client and allows/denies the connection
func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("could not retrieve peer info")
	}

	clientIP := strings.Split(peerInfo.Addr.String(), ":")[0] // Extract IP address
	if !WhitelistedIPs[clientIP] {
		log.Printf("Unauthorized connection attempt from IP: %s", clientIP)
		return nil, fmt.Errorf("unauthorized IP: %s", clientIP)
	}

	return handler(ctx, req)
}
