package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
)

// ExecuteBulk executes a bulk operation with direct inputs
func ExecuteBulk(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}
	orchestrator, hasOrchestrator := middleware.GetBulkOrchestrator(c)
	stateMachineID := c.Param("stateMachineId")

	var req models.ExecuteBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Build inputs from request
	inputs := req.Inputs

	if len(inputs) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "No inputs provided",
			Message: "Provide inputs as a JSON array",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Default values
	if req.NamePrefix == "" {
		req.NamePrefix = fmt.Sprintf("bulk-%d", time.Now().Unix())
	}
	if req.Concurrency == 0 {
		req.Concurrency = 10
	}

	// Build bulk options
	bulkOpts := &statemachine.BulkExecutionOptions{
		NamePrefix:        req.NamePrefix,
		ConcurrentBatches: req.Concurrency,
		StopOnError:       req.StopOnError,
		DoMicroBatch:      req.DoMicroBatch,
		MicroBatchSize:    req.MicroBatchSize,
	}

	var execOpts []statemachine.ExecutionOption

	// Generate batch ID
	batchID := fmt.Sprintf("%s-%d", req.NamePrefix, time.Now().Unix())

	// Execute bulk asynchronously using background context
	// We don't wait for completion - run in background
	go func() {
		bgCtx := context.Background()

		// Load state machine with background context to avoid cancellation
		sm, err := persistent.NewFromDefnId(bgCtx, stateMachineID, repoManager)
		if err != nil {
			fmt.Printf("Failed to load state machine for bulk execution: %v\n", err)
			return
		}

		if hasOrchestrator && orchestrator != nil && req.DoMicroBatch {
			// Use orchestrator for micro-batch lifecycle management
			errorChan, err := orchestrator.RunBulk(bgCtx, batchID, inputs, stateMachineID, bulkOpts, execOpts)
			if err != nil {
				// Log error but don't fail the HTTP request
				fmt.Printf("Bulk execution failed to start: %v\n", err)
				return
			}

			// Monitor for errors in background
			go func() {
				for err := range errorChan {
					fmt.Printf("Bulk execution error: %v\n", err)
				}
			}()
		} else {
			// Fallback to regular bulk execution without orchestration
			_, err := sm.ExecuteBulk(bgCtx, inputs, bulkOpts, execOpts...)
			if err != nil {
				fmt.Printf("Bulk execution failed: %v\n", err)
			}
		}
	}()

	// Return immediately - bulk is running in background
	c.JSON(http.StatusAccepted, models.BulkExecutionResponse{
		OrchestratorID: batchID,
		BatchID:        batchID,
		Status:         "Accepted",
		TotalEnqueued:  len(inputs),
		TotalFailed:    0,
		Mode:           req.Mode,
	})
}

// GetBulkStatus retrieves the status of a bulk execution
func GetBulkStatus(c *gin.Context) {
	orchestrator, ok := middleware.GetBulkOrchestrator(c)
	if !ok || orchestrator == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "BulkOrchestrator not configured",
			Message: "Bulk orchestration is not available",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("orchestratorId")

	// For now, return a basic status response
	// The orchestrator tracks state internally via Redis
	c.JSON(http.StatusOK, models.BulkStatusResponse{
		OrchestratorID: batchID,
		Status:         "Running",
		Progress: &models.BulkProgress{
			TotalBatches:    0,
			CurrentBatch:    0,
			TotalExecutions: 0,
		},
	})
}

// PauseBulkExecution pauses a running bulk execution by signaling the orchestrator
func PauseBulkExecution(c *gin.Context) {
	orchestrator, ok := middleware.GetBulkOrchestrator(c)
	if !ok || orchestrator == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "BulkOrchestrator not configured",
			Message: "Bulk orchestration is not available",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("orchestratorId")

	// Signal pause to the orchestrator
	err := orchestrator.Signal(c.Request.Context(), batchID, "pause", "User requested pause")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to pause bulk execution",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.BulkActionResponse{
		OrchestratorID: batchID,
		Action:         "pause",
		Success:        true,
		Message:        "Pause signal sent successfully",
	})
}

// ResumeBulkExecution resumes a paused bulk execution
func ResumeBulkExecution(c *gin.Context) {
	orchestrator, ok := middleware.GetBulkOrchestrator(c)
	if !ok || orchestrator == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "BulkOrchestrator not configured",
			Message: "Bulk orchestration is not available",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("orchestratorId")

	// Signal resume to the orchestrator
	err := orchestrator.Signal(c.Request.Context(), batchID, "resume", "User requested resume")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to resume bulk execution",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.BulkActionResponse{
		OrchestratorID: batchID,
		Action:         "resume",
		Success:        true,
		Message:        "Resume signal sent successfully",
	})
}

// CancelBulkExecution cancels a bulk execution
func CancelBulkExecution(c *gin.Context) {
	orchestrator, ok := middleware.GetBulkOrchestrator(c)
	if !ok || orchestrator == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "BulkOrchestrator not configured",
			Message: "Bulk orchestration is not available",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("orchestratorId")

	// Signal cancel to the orchestrator
	err := orchestrator.Signal(c.Request.Context(), batchID, "cancel", "User requested cancel")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to cancel bulk execution",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.BulkActionResponse{
		OrchestratorID: batchID,
		Action:         "cancel",
		Success:        true,
		Message:        "Cancel signal sent successfully",
	})
}

// ListBulkExecutions lists all bulk executions
func ListBulkExecutions(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	// List executions with bulk-related filters
	filter := &repository.ExecutionFilter{
		Limit:  100,
		Offset: 0,
	}

	executions, err := repoManager.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to list executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var bulkStatuses []models.BulkStatusResponse
	for _, exec := range executions {
		bulkStatus := models.BulkStatusResponse{
			OrchestratorID: exec.ExecutionID,
			Status:         exec.Status,
		}
		bulkStatuses = append(bulkStatuses, bulkStatus)
	}

	c.JSON(http.StatusOK, gin.H{
		"bulkExecutions": bulkStatuses,
		"total":          len(bulkStatuses),
	})
}
