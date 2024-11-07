package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetGenesisData retrieves the genesis data stored in the database
func GetGenesisData(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var genesisData string
		err := db.QueryRow(`SELECT data FROM genesis_data LIMIT 1`).Scan(&genesisData)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Genesis data not found"})
			return
		}

		var parsedData map[string]interface{}
		if err := json.Unmarshal([]byte(genesisData), &parsedData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse genesis data"})
			return
		}

		c.JSON(http.StatusOK, parsedData)
	}
}
