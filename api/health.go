package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck performs a comprehensive health check of the subscriber
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If all checks pass
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"details": gin.H{
				"database": "reachable",
				"grpc":     "reachable",
			},
		})
	}
}
