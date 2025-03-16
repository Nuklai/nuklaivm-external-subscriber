// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nuklai/nuklaivm-external-subscriber/metrics"
)

// Track metrics for requests
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := newStatusResponseWriter(w)
		next.ServeHTTP(ww, r)

		// Record
		duration := time.Since(start).Seconds()
		endpoint := getRoutePattern(r)
		statusCode := strconv.Itoa(ww.statusCode)

		metrics.ApiRequestsTotal.WithLabelValues(r.Method, endpoint, statusCode).Inc()
		metrics.ApiRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration)
	})
}

// Returns the route pattern from the request
func getRoutePattern(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route == nil {
		return "unknown"
	}

	path, err := route.GetPathTemplate()
	if err != nil {
		return "unknown"
	}

	return path
}

// Wrapper for http.ResponseWriter for status code
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (srw *statusResponseWriter) WriteHeader(code int) {
	srw.statusCode = code
	srw.ResponseWriter.WriteHeader(code)
}
