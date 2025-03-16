// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/nuklai/nuklaivm-external-subscriber/config"
	"github.com/nuklai/nuklaivm-external-subscriber/db/repository"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// API server
type Server struct {
	config     *config.ServerConfig
	repository repository.Repository
	router     *mux.Router
	server     *http.Server
}

func NewServer(cfg *config.ServerConfig, repo repository.Repository) *Server {
	router := mux.NewRouter()

	server := &Server{
		config:     cfg,
		repository: repo,
		router:     router,
	}

	server.setupRoutes()

	return server
}

// API routes
func (s *Server) setupRoutes() {
	// Version API
	apiV1 := s.router.PathPrefix("/api/v1").Subrouter()

	// Health
	apiV1.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Blocks
	apiV1.HandleFunc("/blocks", s.getBlocksHandler).Methods("GET")
	apiV1.HandleFunc("/blocks/{height}", s.getBlockByHeightHandler).Methods("GET")

	// Transactions
	apiV1.HandleFunc("/transactions", s.getTransactionsHandler).Methods("GET")
	apiV1.HandleFunc("/transactions/{hash}", s.getTransactionByHashHandler).Methods("GET")

	// Accounts
	apiV1.HandleFunc("/accounts/{address}", s.getAccountHandler).Methods("GET")
	apiV1.HandleFunc("/accounts/{address}/transactions", s.getAccountTransactionsHandler).Methods("GET")

	// Datasets
	apiV1.HandleFunc("/datasets", s.getDatasetsHandler).Methods("GET")
	apiV1.HandleFunc("/datasets/{denom}", s.getDatasetByDenomHandler).Methods("GET")
	apiV1.HandleFunc("/datasets/{denom}/contributions", s.getDatasetContributionsHandler).Methods("GET")

	// Validators
	apiV1.HandleFunc("/validators", s.getValidatorsHandler).Methods("GET")
	apiV1.HandleFunc("/validators/{address}", s.getValidatorByAddressHandler).Methods("GET")

	// Events
	apiV1.HandleFunc("/events", s.getEventsHandler).Methods("GET")
	apiV1.HandleFunc("/events/types", s.getEventTypesHandler).Methods("GET")

	// Middleware
	s.router.Use(loggingMiddleware)
	s.router.Use(metricsMiddleware)

	if s.config.EnableRateLimiting {
		s.router.Use(rateLimitingMiddleware(s.config.RequestsPerMinute))
	}
}

// Start API server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting API server", zap.String("address", addr))

	var err error
	if s.config.EnableTLS {
		err = s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
	} else {
		err = s.server.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	return nil
}

// Stop API server
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("Stopping API server")

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Log HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

// Rate limiting per IP
func rateLimitingMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	type client struct {
		count     int
		lastSeen  time.Time
		allowance float64
	}

	clients := make(map[string]*client)
	mu := &sync.Mutex{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user IP
			ip := r.RemoteAddr

			mu.Lock()

			c, exists := clients[ip]
			if !exists {
				clients[ip] = &client{
					count:     0,
					lastSeen:  time.Now(),
					allowance: float64(requestsPerMinute),
				}
				c = clients[ip]
			}

			now := time.Now()
			timePassed := now.Sub(c.lastSeen).Seconds()
			c.lastSeen = now

			c.allowance += timePassed * (float64(requestsPerMinute) / 60.0)

			if c.allowance > float64(requestsPerMinute) {
				c.allowance = float64(requestsPerMinute)
			}

			if c.allowance < 1.0 {
				mu.Unlock()
				respondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}

			c.allowance -= 1.0
			c.count++

			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
