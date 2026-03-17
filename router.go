package statemachinegin

import (
	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/handlers"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
)

// SetupRouter sets up the Gin router with all state machine endpoints
func SetupRouter(config *middleware.Config) *gin.Engine {
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.StateMachineMiddleware(config))

	// Base path
	basePath := config.BasePath
	if basePath == "" {
		basePath = "/api"
	}

	api := router.Group(basePath)
	{
		// Health & Monitoring
		api.GET("/health", handlers.HealthCheck)
		// api.GET("/queue/stats", handlers.GetQueueStats)

		// API Documentation
		api.GET("/openapi.json", handlers.GetOpenAPISpec)

		// State Machine Management
		api.POST("/state-machines", handlers.CreateStateMachine)
		api.GET("/state-machines/:stateMachineId", handlers.GetStateMachine)
		api.GET("/state-machines", handlers.ListStateMachines)

		// Execution Management
		api.POST("/state-machines/:stateMachineId/executions", handlers.StartExecution)
		api.GET("/state-machines/:stateMachineId/executions", handlers.ListExecutions)
		api.GET("/transformers", handlers.ListTransformers)
		api.GET("/state-machines/:stateMachineId/executions/count", handlers.CountExecutions)
		api.GET("/executions/:executionId", handlers.GetExecution)
		api.DELETE("/executions/:executionId", handlers.StopExecution)
		api.GET("/executions/:executionId/history", handlers.GetExecutionHistory)

		// Batch Execution
		api.POST("/state-machines/:stateMachineId/executions/batch", handlers.ExecuteBatch)
		api.GET("/batch/:batchId/status", handlers.GetBatchStatus)
		api.POST("/batch/:batchId/pause", handlers.PauseBatchExecution)
		api.POST("/batch/:batchId/resume", handlers.ResumeBatchExecution)
		api.DELETE("/batch/:batchId", handlers.CancelBatchExecution)
		api.GET("/batch", handlers.ListBatches)
		api.POST("/queue/enqueue", handlers.EnqueueExecution)

		// Bulk Execution with Orchestration
		api.POST("/state-machines/:stateMachineId/executions/bulk", handlers.ExecuteBulk)
		api.POST("/state-machines/:stateMachineId/executions/bulk-form", handlers.ExecuteBulkForm)
		api.GET("/bulk/:orchestratorId/status", handlers.GetBulkStatus)
		api.POST("/bulk/:orchestratorId/pause", handlers.PauseBulkExecution)
		api.POST("/bulk/:orchestratorId/resume", handlers.ResumeBulkExecution)
		api.DELETE("/bulk/:orchestratorId", handlers.CancelBulkExecution)
		api.GET("/bulk", handlers.ListBulkExecutions)

		// Message/Resume
		api.POST("/executions/:executionId/resume", handlers.ResumeExecution)
		api.POST("/state-machines/:stateMachineId/resume-by-correlation", handlers.ResumeByCorrelation)
		api.GET("/state-machines/:stateMachineId/waiting", handlers.FindWaitingExecutions)
		api.POST("/orchestrator/resume", handlers.ResumeOrchestrator)
	}

	return router
}

// NewServer creates a new Gin server with state machine routes
func NewServer(config *middleware.Config) *gin.Engine {
	return SetupRouter(config)
}
