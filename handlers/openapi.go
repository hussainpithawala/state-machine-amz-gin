package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// GetOpenAPISpec serves the OpenAPI specification
func GetOpenAPISpec(c *gin.Context) {
	// Try to read from the file system
	// First check current directory, then parent directory
	paths := []string{
		"openapi.json",
		"../openapi.json",
		filepath.Join(".", "openapi.json"),
	}

	var specData []byte
	var err error

	for _, path := range paths {
		specData, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load OpenAPI specification",
			"message": "openapi.json file not found",
		})
		return
	}

	// Parse to validate it's valid JSON
	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse OpenAPI specification",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, spec)
}
