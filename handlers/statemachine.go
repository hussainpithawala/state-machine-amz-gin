package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
)

// CreateStateMachine creates a new state machine
func CreateStateMachine(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Repository manager not configured",
			Message: "Repository manager is not available in context",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var req models.CreateStateMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Convert definition to JSON bytes
	defBytes, err := json.Marshal(req.Definition)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid definition",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Create state machine
	sm, err := persistent.New(defBytes, true, req.ID, repoManager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create state machine",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Save definition
	if err := sm.SaveDefinition(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to save state machine definition",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Retrieve saved definition
	record, err := repoManager.GetStateMachine(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve state machine",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, models.StateMachineResponse{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		Definition:  json.RawMessage(record.Definition),
		Type:        record.Type,
		Version:     record.Version,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		Metadata:    record.Metadata,
	})
}

// GetStateMachine retrieves a state machine by ID
func GetStateMachine(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	stateMachineID := c.Param("stateMachineId")
	record, err := repoManager.GetStateMachine(c.Request.Context(), stateMachineID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "State machine not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.StateMachineResponse{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		Definition:  json.RawMessage(record.Definition),
		Type:        record.Type,
		Version:     record.Version,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		Metadata:    record.Metadata,
	})
}

// ListStateMachines lists all state machines with optional filtering
func ListStateMachines(c *gin.Context) {
	repoManager, ok := middleware.GetRepositoryManager(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Repository manager not configured",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	// Parse query parameters
	name := c.Query("name")

	filter := &repository.DefinitionFilter{
		Name: name,
	}

	records, err := repoManager.ListStateMachines(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to list state machines",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert to response
	stateMachines := make([]*models.StateMachineResponse, len(records))
	for i, record := range records {
		stateMachines[i] = &models.StateMachineResponse{
			ID:          record.ID,
			Name:        record.Name,
			Description: record.Description,
			Definition:  json.RawMessage(record.Definition),
			Type:        record.Type,
			Version:     record.Version,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
			Metadata:    record.Metadata,
		}
	}

	c.JSON(http.StatusOK, models.ListStateMachinesResponse{
		StateMachines: stateMachines,
		Total:         len(stateMachines),
	})
}
