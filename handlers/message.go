package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/batch"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/execution"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
)

// ResumeExecution resumes a paused execution (Message state)
func ResumeExecution(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	executionID := c.Param("executionId")

	var req models.ResumeExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

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

	// Check if execution is paused
	if record.Status != "PAUSED" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Execution is not paused",
			Message: "Only paused executions can be resumed",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Load state machine
	sm, err := persistent.NewFromDefnId(c.Request.Context(), record.StateMachineID, repoManager)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "State machine not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	// Create execution context from record
	execCtx := &execution.Execution{
		ID:             record.ExecutionID,
		StateMachineID: record.StateMachineID,
		Name:           record.Name,
		Status:         record.Status,
		CurrentState:   record.CurrentState,
		Input:          record.Input,
		Output:         req.Output, // Use the output from the resume request
		StartTime:      *record.StartTime,
	}

	// Resume execution
	result, err := sm.ResumeExecution(c.Request.Context(), execCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to resume execution",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.ExecutionResponse{
		ExecutionID:    result.ID,
		StateMachineID: result.StateMachineID,
		Name:           result.Name,
		Status:         result.Status,
		CurrentState:   result.CurrentState,
		Input:          result.Input,
		Output:         result.Output,
		StartTime:      &result.StartTime,
		EndTime:        &result.EndTime,
		Error:          "",
	})
}

// ResumeByCorrelation resumes executions waiting on a correlation key/value
func ResumeByCorrelation(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	stateMachineID := c.Param("stateMachineId")

	var req models.ResumeByCorrelationRequest
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

	// Find waiting executions
	waitingExecutions, err := sm.FindWaitingExecutionsByCorrelation(
		c.Request.Context(),
		req.CorrelationKey,
		req.CorrelationValue,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to find waiting executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if len(waitingExecutions) == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "No waiting executions found",
			Message: "No executions are waiting for this correlation",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Resume each execution
	resumedIDs := make([]string, 0, len(waitingExecutions))
	for _, record := range waitingExecutions {
		execCtx := &execution.Execution{
			ID:             record.ExecutionID,
			StateMachineID: record.StateMachineID,
			Name:           record.Name,
			Status:         record.Status,
			CurrentState:   record.CurrentState,
			Input:          record.Input,
			Output:         req.Output,
			StartTime:      *record.StartTime,
		}

		_, err := sm.ResumeExecution(c.Request.Context(), execCtx)
		if err == nil {
			resumedIDs = append(resumedIDs, record.ExecutionID)
		}
	}

	c.JSON(http.StatusOK, models.ResumeByCorrelationResponse{
		ResumedCount: len(resumedIDs),
		ExecutionIDs: resumedIDs,
	})
}

// FindWaitingExecutions finds executions waiting on a correlation
func FindWaitingExecutions(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	stateMachineID := c.Param("stateMachineId")
	correlationKey := c.Query("correlationKey")
	correlationValue := c.Query("correlationValue")

	if correlationKey == "" || correlationValue == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing parameters",
			Message: "correlationKey and correlationValue are required",
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

	// Find waiting executions
	records, err := sm.FindWaitingExecutionsByCorrelation(
		c.Request.Context(),
		correlationKey,
		correlationValue,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to find waiting executions",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert to response
	executions := make([]*models.ExecutionResponse, len(records))
	for i, record := range records {
		executions[i] = &models.ExecutionResponse{
			ExecutionID:           record.ExecutionID,
			StateMachineID:        record.StateMachineID,
			Name:                  record.Name,
			Status:                record.Status,
			CurrentState:          record.CurrentState,
			Input:                 record.Input,
			Output:                record.Output,
			StartTime:             record.StartTime,
			EndTime:               record.EndTime,
			Error:                 record.Error,
			Metadata:              record.Metadata,
			HistorySequenceNumber: record.HistorySequenceNumber,
		}
	}

	c.JSON(http.StatusOK, executions)
}

// ResumeOrchestrator resumes a stuck orchestrator (WaitForMicroBatchCompletion/WaitForBulkMicroBatchCompletion)
// to handle exceptional conditions. This is useful when the orchestrator is waiting for micro-batch completion
// signals that may have been lost or delayed.
func ResumeOrchestrator(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	queueClient, ok := middleware.GetQueueClient(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Queue client not configured",
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

	var req models.ResumeOrchestratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate orchestrator SM ID
	if req.OrchestratorSMID != batch.OrchestratorStateMachineID && req.OrchestratorSMID != batch.BulkOrchestratorStateMachineID {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid orchestratorSmId",
			Message: fmt.Sprintf("Must be either '%s' or '%s'", batch.OrchestratorStateMachineID, batch.BulkOrchestratorStateMachineID),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Create orchestrator instance
	parentSM, err := persistent.New(batch.OrchestratorDefinitionJSON(), true, batch.OrchestratorStateMachineID, repoManager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create parent state machine",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}
	parentSM.SetQueueClient(queueClient)

	smFactory := func(ctx context.Context, smID string, manager *repository.Manager) (batch.StateMachine, error) {
		return persistent.NewFromDefnId(ctx, smID, manager)
	}

	smCreator := func(def []byte, isJSON bool, smID string, manager *repository.Manager) (batch.StateMachine, error) {
		return persistent.New(def, isJSON, smID, manager)
	}

	orchestrator, err := batch.NewOrchestrator(c.Request.Context(), redisClient, parentSM, smFactory, smCreator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create orchestrator",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Create MicroBatchMeta from request
	mbMeta := batch.MicroBatchMeta{
		BatchID:          req.BatchID,
		MicroBatchID:     req.MicroBatchID,
		OrchestratorSMID: req.OrchestratorSMID,
	}

	// Resume the orchestrator
	err = orchestrator.ResumeOrchestrator(c.Request.Context(), mbMeta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to resume orchestrator",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.ResumeOrchestratorResponse{
		BatchID:      req.BatchID,
		MicroBatchID: req.MicroBatchID,
		Status:       "resumed",
		Message:      "Orchestrator resumed successfully",
	})
}
