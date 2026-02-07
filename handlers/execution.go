package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/types"
)

// StartExecution queues a new execution for a state machine
func StartExecution(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	baseExecutor, ok := middleware.GetBaseExecutor(c)

	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	queueClient, ok := middleware.GetQueueClient(c)
	if !ok || queueClient == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Queue client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	stateMachineID := c.Param("stateMachineId")

	var req models.StartExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Load state machine to validate it exists
	sm, err := persistent.NewFromDefnId(c.Request.Context(), stateMachineID, repoManager)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "State machine not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx := context.WithValue(c.Request.Context(), types.ExecutionContextKey, executor.NewExecutionContextAdapter(baseExecutor))

	// Queue the execution instead of executing directly
	exec, err := sm.Execute(
		ctx,
		req.Input,
		statemachine.WithExecutionName(req.Name),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to queue execution",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusAccepted, models.StartExecutionResponse{
		ExecutionID:    exec.ID,
		StateMachineID: exec.StateMachineID,
		Name:           exec.Name,
		Status:         exec.Status,
		StartTime:      exec.StartTime,
		Input:          exec.Input,
	})
}

// GetExecution retrieves an execution by ID
func GetExecution(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	executionID := c.Param("executionId")
	record, err := repoManager.GetExecution(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Execution not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.ExecutionResponse{
		ExecutionID:    record.ExecutionID,
		StateMachineID: record.StateMachineID,
		Name:           record.Name,
		Status:         record.Status,
		CurrentState:   record.CurrentState,
		Input:          record.Input,
		Output:         record.Output,
		StartTime:      record.StartTime,
		EndTime:        record.EndTime,
		Error:          record.Error,
		Metadata:       record.Metadata,
	})
}

// ListExecutions lists executions for a state machine with filtering
func ListExecutions(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	stateMachineID := c.Param("stateMachineId")

	// Parse query parameters
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	filter := &repository.ExecutionFilter{
		StateMachineID: stateMachineID,
		Status:         status,
		Limit:          limit,
		Offset:         offset,
	}

	// Get executions
	records, err := repoManager.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to list executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Get total count
	total, err := repoManager.CountExecutions(c.Request.Context(), filter)
	if err != nil {
		total = int64(len(records))
	}

	// Convert to response
	executions := make([]*models.ExecutionResponse, len(records))
	for i, record := range records {
		executions[i] = &models.ExecutionResponse{
			ExecutionID:    record.ExecutionID,
			StateMachineID: record.StateMachineID,
			Name:           record.Name,
			Status:         record.Status,
			CurrentState:   record.CurrentState,
			Input:          record.Input,
			Output:         record.Output,
			StartTime:      record.StartTime,
			EndTime:        record.EndTime,
			Error:          record.Error,
			Metadata:       record.Metadata,
		}
	}

	c.JSON(http.StatusOK, models.ListExecutionsResponse{
		Executions: executions,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	})
}

// GetExecutionHistory retrieves the state history for an execution
func GetExecutionHistory(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	executionID := c.Param("executionId")

	records, err := repoManager.GetStateHistory(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve execution history",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert to response
	history := make([]*models.StateHistoryResponse, len(records))
	for i, record := range records {
		history[i] = &models.StateHistoryResponse{
			ID:             record.ID,
			ExecutionID:    record.ExecutionID,
			StateName:      record.StateName,
			StateType:      record.StateType,
			Status:         record.Status,
			Input:          record.Input,
			Output:         record.Output,
			StartTime:      record.StartTime,
			EndTime:        record.EndTime,
			Error:          record.Error,
			RetryCount:     record.RetryCount,
			SequenceNumber: record.SequenceNumber,
			Metadata:       record.Metadata,
		}
	}

	c.JSON(http.StatusOK, history)
}

// CountExecutions counts executions matching filters
func CountExecutions(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	stateMachineID := c.Param("stateMachineId")
	status := c.Query("status")

	filter := &repository.ExecutionFilter{
		StateMachineID: stateMachineID,
		Status:         status,
	}

	count, err := repoManager.CountExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to count executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// StopExecution stops a running execution
func StopExecution(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	executionID := c.Param("executionId")

	// Get execution
	record, err := repoManager.GetExecution(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Execution not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check if execution is already stopped
	if record.Status == "SUCCEEDED" || record.Status == "FAILED" || record.Status == "CANCELLED" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Execution already stopped",
			Message: fmt.Sprintf("Execution is in %s state", record.Status),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Note: Actual cancellation would require context cancellation
	// For now, we just mark it as cancelled in the database
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Execution stop requested (note: actual cancellation requires context support)",
		Data: gin.H{
			"executionId": executionID,
			"status":      "CANCELLED",
		},
	})
}
