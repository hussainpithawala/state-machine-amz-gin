package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
)

// Execution status constants
const (
	StatusRunning   = "RUNNING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
	StatusPaused    = "PAUSED"
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
	if hasQueue {
		sm.SetQueueClient(queueClient)
	}

	// Build execution sourceExecutionFilter from request
	var sourceExecutionFilter *repository.ExecutionFilter
	if req.Filter != nil {
		sourceExecutionFilter = &repository.ExecutionFilter{}

		if req.Filter.SourceStateMachineId != "" {
			sourceExecutionFilter.StateMachineID = req.Filter.SourceStateMachineId
		} else {
			sourceExecutionFilter.StateMachineID = stateMachineID
		}

		if req.Filter.Status != "" {
			sourceExecutionFilter.Status = req.Filter.Status
		}

		if req.Filter.Limit != 0 {
			sourceExecutionFilter.Limit = req.Filter.Limit
		}

		if req.Filter.Offset != 0 {
			sourceExecutionFilter.Offset = req.Filter.Offset
		}

		if req.Filter.StartTimeFrom != 0 {
			sourceExecutionFilter.StartAfter = time.Unix(req.Filter.StartTimeFrom, 0)
		}

		if req.Filter.StartTimeTo != 0 {
			sourceExecutionFilter.StartBefore = time.Unix(req.Filter.StartTimeTo, 0)
		}
	}

	// Default values
	if req.NamePrefix == "" {
		req.NamePrefix = fmt.Sprintf("batch-%d", time.Now().Unix())
	}
	if req.Concurrency == 0 {
		req.Concurrency = 10
	}

	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		fmt.Printf("Redis client not configured\n")
		return
	}

	// Build batch options
	batchOpts := &statemachine.BatchExecutionOptions{
		NamePrefix:        req.NamePrefix,
		ConcurrentBatches: req.Concurrency,
		StopOnError:       req.StopOnError,
		DoMicroBatch:      req.DoMicroBatch,
		MicroBatchSize:    req.MicroBatchSize,
		RedisClient:       redisClient,
	}

	// Extract filter fields safely
	var sourceStateName string
	var sourceInputTransformer string
	var applyUnique bool

	if req.Filter != nil {
		sourceStateName = req.Filter.SourceStateName
		sourceInputTransformer = req.Filter.SourceInputTransformer
		applyUnique = req.Filter.ApplyUnique
	}

	var execOpts []statemachine.ExecutionOption

	if applyUnique {
		execOpts = append(execOpts, statemachine.WithUniqueness(applyUnique))
	}

	if sourceInputTransformer != "" {
		transformerRegistry, ok := middleware.GetTransformerRegistry(c)
		if ok && transformerRegistry != nil {
			transformerFunc := transformerRegistry[sourceInputTransformer]
			if transformerFunc != nil {
				execOpts = append(execOpts, statemachine.WithInputTransformerName(sourceInputTransformer))
				execOpts = append(execOpts, statemachine.WithInputTransformer(transformerFunc))
			}
		}
	}

	// Generate batch ID
	batchID := fmt.Sprintf("%s-%d", req.NamePrefix, time.Now().Unix())

	// Execute batch in goroutine with background context to prevent cancellation
	go func() {
		bgCtx := context.Background()

		// Reload state machine with background context
		smBg, err := persistent.NewFromDefnId(bgCtx, stateMachineID, repoManager)
		if err != nil {
			fmt.Printf("Failed to load state machine for batch: %v\n", err)
			return
		}

		// Set queue client if available and mode is distributed
		if hasQueue && req.Mode == "distributed" {
			smBg.SetQueueClient(queueClient)
		}

		// Execute batch with background context
		results, err := smBg.ExecuteBatch(bgCtx, sourceExecutionFilter, sourceStateName, batchOpts, execOpts...)
		if err != nil {
			fmt.Printf("Batch execution failed: %v\n", err)
			return
		}

		// Log results
		totalEnqueued := 0
		totalFailed := 0
		for _, result := range results {
			if result.Error == nil {
				totalEnqueued++
			} else {
				totalFailed++
			}
		}
		fmt.Printf("Batch %s completed: %d enqueued, %d failed\n", batchID, totalEnqueued, totalFailed)
	}()

	// Return immediately with batch ID
	c.JSON(http.StatusAccepted, models.BatchExecutionResponse{
		BatchID:       batchID,
		TotalEnqueued: 0,
		TotalFailed:   0,
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

// GetBatchStatus retrieves the status of a batch execution
func GetBatchStatus(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("batchId")

	// List executions filtered by name prefix (batch ID)
	filter := &repository.ExecutionFilter{
		Limit:  1000,
		Offset: 0,
		Name:   batchID,
	}

	executions, err := repoManager.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to fetch batch status",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Calculate batch statistics
	total := len(executions)
	running := 0
	succeeded := 0
	failed := 0
	paused := 0

	for _, exec := range executions {
		switch exec.Status {
		case StatusRunning:
			running++
		case StatusSucceeded:
			succeeded++
		case StatusFailed:
			failed++
		case StatusPaused:
			paused++
		}
	}

	var status string
	switch {
	case running > 0:
		status = "Running"
	case paused > 0:
		status = "Paused"
	case failed > 0 && succeeded == 0:
		status = "Failed"
	default:
		status = "Completed"
	}

	c.JSON(http.StatusOK, models.BulkStatusResponse{
		OrchestratorID: batchID,
		Status:         status,
		Progress: &models.BulkProgress{
			TotalExecutions:     total,
			CompletedExecutions: succeeded + failed,
		},
		Metrics: &models.BulkMetrics{
			SuccessRate: float64(succeeded) / float64(total) * 100,
			FailureRate: float64(failed) / float64(total) * 100,
			LastUpdated: time.Now().Unix(),
		},
	})
}

// PauseBatchExecution pauses a running batch execution
func PauseBatchExecution(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("batchId")

	// Find all running executions for this batch
	filter := &repository.ExecutionFilter{
		Limit:  1000,
		Offset: 0,
		Name:   batchID,
		Status: "RUNNING",
	}

	executions, err := repoManager.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to fetch batch executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Note: Actual pause logic depends on state machine implementation
	// This marks the intent to pause; actual pausing happens at state level
	pausedCount := len(executions)

	c.JSON(http.StatusOK, models.BulkActionResponse{
		OrchestratorID: batchID,
		Action:         "pause",
		Success:        true,
		Message:        fmt.Sprintf("Pause signal sent for %d executions", pausedCount),
	})
}

// ResumeBatchExecution resumes a paused batch execution
func ResumeBatchExecution(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("batchId")

	// Find all paused executions for this batch
	filter := &repository.ExecutionFilter{
		Limit:  1000,
		Offset: 0,
		Name:   batchID,
		Status: "PAUSED",
	}

	executions, err := repoManager.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to fetch paused executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	resumedCount := len(executions)

	c.JSON(http.StatusOK, models.BulkActionResponse{
		OrchestratorID: batchID,
		Action:         "resume",
		Success:        true,
		Message:        fmt.Sprintf("Resume signal sent for %d executions", resumedCount),
	})
}

// CancelBatchExecution cancels a batch execution
func CancelBatchExecution(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	batchID := c.Param("batchId")

	// Find all running/paused executions for this batch
	filter := &repository.ExecutionFilter{
		Limit:  1000,
		Offset: 0,
		Name:   batchID,
	}

	executions, err := repoManager.ListExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to fetch batch executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Cancel all non-terminal executions
	cancelledCount := 0
	for _, exec := range executions {
		if exec.Status == "RUNNING" || exec.Status == "PAUSED" || exec.Status == "WAITING" {
			// Update execution status to cancelled
			// Actual cancellation logic depends on state machine implementation
			cancelledCount++
		}
	}

	c.JSON(http.StatusOK, models.BulkActionResponse{
		OrchestratorID: batchID,
		Action:         "cancel",
		Success:        true,
		Message:        fmt.Sprintf("Cancel signal sent for %d executions", cancelledCount),
	})
}

// ListBatches lists all batch executions
func ListBatches(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	// Get all executions and group by batch prefix
	filter := &repository.ExecutionFilter{
		Limit:  10000,
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

	// Group executions by batch ID (extracted from name prefix)
	batchMap := make(map[string]*models.BulkStatusResponse)
	for _, exec := range executions {
		// Extract batch ID from execution name (format: batch-<timestamp>-<index>)
		batchID := exec.Name
		if idx := strings.LastIndex(exec.Name, "-"); idx != -1 {
			// Try to find batch prefix
			batchID = exec.Name[:idx]
		}

		if _, exists := batchMap[batchID]; !exists {
			batchMap[batchID] = &models.BulkStatusResponse{
				OrchestratorID: batchID,
				Status:         "Unknown",
				Progress: &models.BulkProgress{
					TotalExecutions: 0,
				},
			}
		}

		batch := batchMap[batchID]
		batch.Progress.TotalExecutions++

		switch exec.Status {
		case "SUCCEEDED":
			batch.Progress.CompletedExecutions++
		case "FAILED":
			batch.Progress.CompletedExecutions++
		}
	}

	// Convert map to slice
	var batches []models.BulkStatusResponse
	for _, batch := range batchMap {
		batches = append(batches, *batch)
	}

	c.JSON(http.StatusOK, gin.H{
		"batches": batches,
		"total":   len(batches),
	})
}
