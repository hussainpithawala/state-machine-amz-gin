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
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "No Transformer Registry found",
			Code:  http.StatusInternalServerError,
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
