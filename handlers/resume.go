package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/batch"
)

// SignalResume signals a paused batch to resume via Redis signaling.
// This is used for human-in-the-loop pause/resume cycle.
// The operator sets a resume signal that the batch execution polls.
func SignalResume(c *gin.Context) {
	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	var req models.SignalResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Create ResumeController and signal resume
	controller := batch.NewResumeController(redisClient)
	err := controller.Signal(c.Request.Context(), req.BatchID, req.Operator, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to signal resume",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SignalResumeResponse{
		BatchID:   req.BatchID,
		Status:    "signaled",
		ResumedBy: req.Operator,
		ResumedAt: time.Now().UTC().Format(time.RFC3339),
		Message:   "Resume signal set successfully",
	})
}

// RevokeResume removes an unconsumed resume signal (e.g., operator changed their mind).
// This is useful when an operator wants to cancel a previously set resume signal
// before the batch execution has consumed it.
func RevokeResume(c *gin.Context) {
	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	var req models.RevokeResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Create ResumeController and revoke signal
	controller := batch.NewResumeController(redisClient)
	err := controller.Revoke(c.Request.Context(), req.BatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to revoke resume signal",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.RevokeResumeResponse{
		BatchID: req.BatchID,
		Status:  "revoked",
		Message: "Resume signal revoked successfully",
	})
}

// CheckResume checks for a resume signal and atomically consumes it.
// This uses Redis GETDEL to ensure the signal is consumed only once.
// Returns shouldResume=true if a valid signal was present.
func CheckResume(c *gin.Context) {
	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	var req models.CheckResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Create ResumeController and check for signal
	controller := batch.NewResumeController(redisClient)
	result, err := controller.Check(c.Request.Context(), req.BatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to check resume signal",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := models.CheckResumeResponse{
		BatchID:       req.BatchID,
		ShouldResume:  result.ShouldResume,
		SignalPresent: result.ShouldResume,
	}

	if result.ShouldResume {
		response.ResumedBy = result.ResumedBy
		if result.ResumedAt != nil {
			response.ResumedAt = result.ResumedAt.Format(time.RFC3339)
		}
	}

	c.JSON(http.StatusOK, response)
}

// SignalResumeParam is a convenience handler that accepts batchID as a path parameter
// This is useful for simple curl commands or direct API calls
func SignalResumeParam(c *gin.Context) {
	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("batchId")
	operator := c.Query("operator")
	notes := c.Query("notes")

	if operator == "" {
		operator = "unknown"
	}

	// Create ResumeController and signal resume
	controller := batch.NewResumeController(redisClient)
	err := controller.Signal(c.Request.Context(), batchID, operator, notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to signal resume",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SignalResumeResponse{
		BatchID:   batchID,
		Status:    "signaled",
		ResumedBy: operator,
		ResumedAt: time.Now().UTC().Format(time.RFC3339),
		Message:   fmt.Sprintf("Resume signal set successfully for batch %s", batchID),
	})
}

// RevokeResumeParam is a convenience handler that accepts batchID as a path parameter
func RevokeResumeParam(c *gin.Context) {
	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("batchId")

	// Create ResumeController and revoke signal
	controller := batch.NewResumeController(redisClient)
	err := controller.Revoke(c.Request.Context(), batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to revoke resume signal",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.RevokeResumeResponse{
		BatchID: batchID,
		Status:  "revoked",
		Message: fmt.Sprintf("Resume signal revoked successfully for batch %s", batchID),
	})
}

// CheckResumeParam is a convenience handler that accepts batchID as a path parameter
func CheckResumeParam(c *gin.Context) {
	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("batchId")

	// Create ResumeController and check for signal
	controller := batch.NewResumeController(redisClient)
	result, err := controller.Check(c.Request.Context(), batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to check resume signal",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := models.CheckResumeResponse{
		BatchID:       batchID,
		ShouldResume:  result.ShouldResume,
		SignalPresent: result.ShouldResume,
	}

	if result.ShouldResume {
		response.ResumedBy = result.ResumedBy
		if result.ResumedAt != nil {
			response.ResumedAt = result.ResumedAt.Format(time.RFC3339)
		}
	}

	c.JSON(http.StatusOK, response)
}
