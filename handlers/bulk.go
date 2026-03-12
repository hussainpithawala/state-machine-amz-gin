package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}
	queueClient, okQueue := middleware.GetQueueClient(c)
	if !okQueue {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Queue client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

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
		RedisClient:       redisClient,
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

		sm.SetQueueClient(queueClient)
		_, bulkExecuteError := sm.ExecuteBulk(bgCtx, inputs, bulkOpts, execOpts...)
		if bulkExecuteError != nil {
			fmt.Printf("Bulk execution failed: %v\n", err)
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

// ExecuteBulkForm executes a bulk operation with inputs provided as a JSON file via form-data
// Form fields:
// - inputs: JSON file containing array of inputs (required)
// - namePrefix: Prefix for execution names (optional, default: "bulk-{timestamp}")
// - concurrency: Number of concurrent executions (optional, default: 10)
// - mode: Execution mode - "distributed", "concurrent", "sequential" (optional)
// - stopOnError: Stop if an error occurs (optional, default: false)
// - doMicroBatch: Enable micro-batch processing (optional, default: false)
// - microBatchSize: Size of each micro-batch (optional, default: 100)
// - orchestratorId: Custom orchestrator ID (optional)
// - pauseThreshold: Failure rate threshold for auto-pause 0.0-1.0 (optional)
// - resumeStrategy: Resume strategy - "manual", "automatic", "timeout" (optional)
// - timeoutSeconds: Timeout for automatic resume (optional)
func ExecuteBulkForm(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}
	stateMachineID := c.Param("stateMachineId")

	redisClient, ok := middleware.GetRedisClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Redis client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	queueClient, okQueue := middleware.GetQueueClient(c)
	if !okQueue {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Queue client not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	// Parse form-data with memory limit of 32MB
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid form data",
			Message: fmt.Sprintf("Failed to parse form: %v", err),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get inputs file
	file, header, err := c.Request.FormFile("inputs")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing inputs file",
			Message: "Please provide a JSON file containing the inputs array",
			Code:    http.StatusBadRequest,
		})
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}
	}()

	// Validate file size (max 10MB)
	if header.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "File too large",
			Message: "Inputs file must be less than 10MB",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Read and parse JSON file
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to read file",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse JSON array
	var inputs []interface{}
	if err := json.Unmarshal(fileBytes, &inputs); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid JSON format",
			Message: fmt.Sprintf("Inputs file must contain a valid JSON array: %v", err),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if len(inputs) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "No inputs provided",
			Message: "Inputs array must contain at least one item",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Parse form fields
	namePrefix := c.PostForm("namePrefix")
	concurrencyStr := c.PostForm("concurrency")
	mode := c.PostForm("mode")
	stopOnErrorStr := c.PostForm("stopOnError")
	doMicroBatchStr := c.PostForm("doMicroBatch")
	microBatchSizeStr := c.PostForm("microBatchSize")
	orchestratorID := c.PostForm("orchestratorId")

	// Set defaults
	if namePrefix == "" {
		namePrefix = fmt.Sprintf("bulk-%d", time.Now().Unix())
	}

	concurrency := 10
	if concurrencyStr != "" {
		if parsed, err := strconv.Atoi(concurrencyStr); err == nil && parsed > 0 {
			concurrency = parsed
		}
	}

	stopOnError := false
	if stopOnErrorStr != "" {
		stopOnError, _ = strconv.ParseBool(stopOnErrorStr)
	}

	doMicroBatch := false
	if doMicroBatchStr != "" {
		doMicroBatch, _ = strconv.ParseBool(doMicroBatchStr)
	}

	microBatchSize := 100
	if microBatchSizeStr != "" {
		if parsed, err := strconv.Atoi(microBatchSizeStr); err == nil && parsed > 0 {
			microBatchSize = parsed
		}
	}

	// Build bulk options
	bulkOpts := &statemachine.BulkExecutionOptions{
		NamePrefix:        namePrefix,
		ConcurrentBatches: concurrency,
		StopOnError:       stopOnError,
		DoMicroBatch:      doMicroBatch,
		MicroBatchSize:    microBatchSize,
		RedisClient:       redisClient,
	}

	var execOpts []statemachine.ExecutionOption

	// Generate batch ID
	batchID := fmt.Sprintf("%s-%d", namePrefix, time.Now().Unix())
	if orchestratorID != "" {
		batchID = orchestratorID
	}

	// Execute bulk asynchronously using background context
	go func() {
		bgCtx := context.Background()

		// Load state machine with background context to avoid cancellation
		sm, err := persistent.NewFromDefnId(bgCtx, stateMachineID, repoManager)
		if err != nil {
			fmt.Printf("Failed to load state machine for bulk execution: %v\n", err)
			return
		}

		sm.SetQueueClient(queueClient)

		_, bulkError := sm.ExecuteBulk(bgCtx, inputs, bulkOpts, execOpts...)
		if bulkError != nil {
			fmt.Printf("Bulk execution failed: %v\n", bulkError)
		}
	}()

	// Return immediately - bulk is running in background
	c.JSON(http.StatusAccepted, models.BulkExecutionResponse{
		OrchestratorID: batchID,
		BatchID:        batchID,
		Status:         "Accepted",
		TotalEnqueued:  len(inputs),
		TotalFailed:    0,
		Mode:           mode,
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
