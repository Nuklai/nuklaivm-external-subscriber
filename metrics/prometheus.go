// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package metrics

import (
	"fmt"
	"net/http"

	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	// Track the number of blocks processed
	BlocksProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nuklai_indexer_blocks_processed_total",
			Help: "The total number of blocks processed",
		},
	)

	// Track the time it takes to process a block
	BlockProcessingTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "nuklai_indexer_block_processing_seconds",
			Help:    "Time taken to process a block",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
		},
	)

	// Track the number of transactions processed
	TransactionsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nuklai_indexer_transactions_processed_total",
			Help: "The total number of transactions processed",
		},
	)

	// Track the number of events processed
	EventsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nuklai_indexer_events_processed_total",
			Help: "The total number of events processed",
		},
	)

	// Track the current sync height
	SyncHeight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "nuklai_indexer_sync_height",
			Help: "The current sync height",
		},
	)

	// Track the latest known block height
	LatestKnownHeight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "nuklai_indexer_latest_known_height",
			Help: "The latest known block height",
		},
	)

	// Track the number of database connections
	DatabaseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "nuklai_indexer_database_connections",
			Help: "The number of database connections",
		},
	)

	// Track the total number of API requests
	ApiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nuklai_indexer_api_requests_total",
			Help: "The total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Track the API request latency
	ApiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nuklai_indexer_api_request_duration_seconds",
			Help:    "The API request latency in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"method", "endpoint"},
	)
)

// Register with Prometheus
func InitMetrics() {
	prometheus.MustRegister(BlocksProcessed)
	prometheus.MustRegister(BlockProcessingTime)
	prometheus.MustRegister(TransactionsProcessed)
	prometheus.MustRegister(EventsProcessed)
	prometheus.MustRegister(SyncHeight)
	prometheus.MustRegister(LatestKnownHeight)
	prometheus.MustRegister(DatabaseConnections)
	prometheus.MustRegister(ApiRequestsTotal)
	prometheus.MustRegister(ApiRequestDuration)
}

func StartMetricsServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", port)

	logger.Info("Starting metrics server", zap.String("address", addr))
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("Failed to start metrics server", zap.Error(err))
		}
	}()
}
