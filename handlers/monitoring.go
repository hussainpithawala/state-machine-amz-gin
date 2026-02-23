package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
)

// HealthCheck performs a health check
func HealthCheck(c *gin.Context) {
	repoManager, hasRepo := middleware.GetRepositoryManager(c)
	_, hasQueue := middleware.GetQueueClient(c)

	services := make(map[string]string)

	// Check repository
	if hasRepo {
		// Try a simple ping operation
		_, err := repoManager.ListStateMachines(c.Request.Context(), nil)
		if err != nil {
			services["database"] = "down"
		} else {
			services["database"] = "up"
		}
	} else {
		services["database"] = "not_configured"
	}

	// Check queue
	if hasQueue {
		services["queue"] = "up"
	} else {
		services["queue"] = "not_configured"
	}

	// Determine overall status
	status := "healthy"
	for _, serviceStatus := range services {
		if serviceStatus == "down" {
			status = "unhealthy"
			break
		}
	}

	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, models.HealthResponse{
		Status:   status,
		Services: services,
	})
}

// GetQueueStats retrieves queue statistics
// func GetQueueStats(c *gin.Context) {
// 	queueClient, ok := middleware.GetQueueClient(c)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
// 			Error:   "Queue client not configured",
// 			Message: "Distributed queue is not available",
// 			Code:    http.StatusInternalServerError,
// 		})
// 		return
// 	}
//
// 	// Get queue statistics from Redis
// 	stats, err := queueClient.GetStats()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
// 			Error:   "Failed to retrieve queue stats",
// 			Message: err.Error(),
// 			Code:    http.StatusInternalServerError,
// 		})
// 		return
// 	}
//
// 	// Convert to response format
// 	queueStats := make(map[string]models.QueueStats)
// 	for queueName, stat := range stats {
// 		queueStats[queueName] = models.QueueStats{
// 			Pending: stat.Pending,
// 			Active:  stat.Active,
// 			Failed:  stat.Failed,
// 		}
// 	}
//
//	c.JSON(http.StatusOK, models.QueueStatsResponse{
//		Queues: queueStats,
//	})
//}
