package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
)

// ListExecutions lists executions for a state machine with filtering
func ListTransformers(c *gin.Context) {
	transformerRegistry, ok := middleware.GetTransformerRegistry(c)
	if !ok || transformerRegistry == nil {
		c.JSON(http.StatusOK, models.ListTransformersResponse{
			Transformers: []string{},
		})
		return
	}

	// Get keys
	keys := make([]string, 0, len(transformerRegistry))
	for k := range transformerRegistry {
		keys = append(keys, k)
	}
	c.JSON(http.StatusOK, models.ListTransformersResponse{
		Transformers: keys,
	})
}
