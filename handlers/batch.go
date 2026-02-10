package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
)

// ExecuteBatch executes a batch of executions
func ExecuteBatch(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	queueClient, hasQueue := middleware.GetQueueClient(c)
	stateMachineID := c.Param("stateMachineId")

	var req models.ExecuteBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Load state machine
	sm, err := persistent.NewFromDefnId(c.Request.Context(), stateMachineID, repoManager)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "State machine not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	// Set queue client if available and mode is distributed
	if hasQueue && req.Mode == "distributed" {
		sm.SetQueueClient(queueClient)
	}

	// Build execution filter from request
	var filter *repository.ExecutionFilter
	if req.Filter != nil {
		filter = &repository.ExecutionFilter{}

		if req.Filter.SourceStateMachineId != "" {
			filter.StateMachineID = req.Filter.SourceStateMachineId
		} else {
			filter.StateMachineID = stateMachineID
		}

		if req.Filter.Status != "" {
			filter.Status = req.Filter.Status
		}

		if req.Filter.Limit != 0 {
			filter.Limit = req.Filter.Limit
		}

		if req.Filter.Offset != 0 {
			filter.Offset = req.Filter.Offset
		}

		if req.Filter.StartTimeFrom != 0 {
			filter.StartAfter = time.Unix(req.Filter.StartTimeFrom, 0)
		}

		if req.Filter.StartTimeTo != 0 {
			filter.StartBefore = time.Unix(req.Filter.StartTimeTo, 0)
		}

	}

	// Default values
	if req.NamePrefix == "" {
		req.NamePrefix = fmt.Sprintf("batch-%d", time.Now().Unix())
	}
	if req.Concurrency == 0 {
		req.Concurrency = 10
	}

	// Build batch options
	batchOpts := &statemachine.BatchExecutionOptions{
		NamePrefix:        req.NamePrefix,
		ConcurrentBatches: req.Concurrency,
		StopOnError:       req.StopOnError,
		//Mode:              req.Mode, // "distributed", "concurrent", "sequential"
	}

	// Execute batch
	results, err := sm.ExecuteBatch(c.Request.Context(), filter, "", batchOpts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Batch execution failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Count successes and failures
	totalEnqueued := 0
	totalFailed := 0
	for _, result := range results {
		if result.Error == nil {
			totalEnqueued++
		} else {
			totalFailed++
		}
	}

	batchID := fmt.Sprintf("%s-%d", req.NamePrefix, time.Now().Unix())

	c.JSON(http.StatusOK, models.BatchExecutionResponse{
		BatchID:       batchID,
		TotalEnqueued: totalEnqueued,
		TotalFailed:   totalFailed,
		Mode:          req.Mode,
	})
}

// EnqueueExecution enqueues an execution task to the distributed queue
func EnqueueExecution(c *gin.Context) {
	queueClient, ok := middleware.GetQueueClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Queue client not configured",
			Message: "Distributed queue is not available",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var req models.EnqueueExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Create task payload
	payload := &queue.ExecutionTaskPayload{
		StateMachineID:    req.StateMachineID,
		ExecutionName:     req.ExecutionName,
		Input:             req.Input,
		SourceExecutionID: req.SourceExecutionID,
		SourceStateName:   req.SourceStateName,
	}

	// Enqueue the task
	taskInfo, err := queueClient.EnqueueExecution(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to enqueue execution",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, models.EnqueueExecutionResponse{
		TaskID:     taskInfo.ID,
		Queue:      taskInfo.Queue,
		EnqueuedAt: time.Now(),
	})
}
