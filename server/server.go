// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package server

import (
	"context"
	"database/sql"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/hypersdk/chain"
	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var mu = &sync.Mutex{}

// Server implements the ExternalSubscriberServer
type Server struct {
	pb.UnimplementedExternalSubscriberServer
	db     *sql.DB
	parser chain.Parser
}

// StartGRPCServer starts the gRPC server for receiving block data
func StartGRPCServer(db *sql.DB, port string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in StartGRPCServer: %v", r)
		}
	}()

	loadWhitelist()

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Failed to listen on port %s: %v", port, err)
		return err
	}

	serverOptions := []grpc.ServerOption{
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(UnaryInterceptor),
	}
	grpcServer := grpc.NewServer(serverOptions...)
	pb.RegisterExternalSubscriberServer(grpcServer, &Server{db: db})
	reflection.Register(grpcServer)

	log.Printf("External Subscriber server is listening on port %s...\n", port)
	return grpcServer.Serve(lis)
}

// StartGRPCServerWithRetries retries gRPC server startup in case of failure
func StartGRPCServerWithRetries(db *sql.DB, port string, retries int) {
	for i := 0; i < retries; i++ {
		err := StartGRPCServer(db, port)
		if err != nil {
			log.Printf("gRPC server failed to start: %v. Retrying (%d/%d)...", err, i+1, retries)
			time.Sleep(5 * time.Second)
			continue
		}
		return
	}
	log.Fatal("gRPC server failed to start after maximum retries")
}

// Initialize receives genesis data for initialization and saves it to the database
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in Initialize: %v", r)
		}
	}()

	mu.Lock()
	defer mu.Unlock()

	parser, err := handleInitialize(s.db, req)
	if err != nil {
		log.Printf("Error initializing External Subscriber: %v", err)
		return nil, err
	}
	s.parser = parser
	return &emptypb.Empty{}, nil
}

// AcceptBlock processes a new block
func (s *Server) AcceptBlock(ctx context.Context, req *pb.BlockRequest) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in AcceptBlock: %v", r)
		}
	}()
	mu.Lock()
	defer mu.Unlock()

	err := handleAcceptBlock(s.db, s.parser, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
