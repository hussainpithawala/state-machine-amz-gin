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
		basePath = "/api/v1"
	}

	api := router.Group(basePath)
	{
		// Health & Monitoring
		api.GET("/health", handlers.HealthCheck)
		//api.GET("/queue/stats", handlers.GetQueueStats)

		// State Machine Management
		api.POST("/state-machines", handlers.CreateStateMachine)
		api.GET("/state-machines/:stateMachineId", handlers.GetStateMachine)
		api.GET("/state-machines", handlers.ListStateMachines)

		// Execution Management
		api.POST("/state-machines/:stateMachineId/executions", handlers.StartExecution)
		api.GET("/state-machines/:stateMachineId/executions", handlers.ListExecutions)
		api.GET("/state-machines/:stateMachineId/executions/count", handlers.CountExecutions)
		api.GET("/executions/:executionId", handlers.GetExecution)
		api.DELETE("/executions/:executionId", handlers.StopExecution)
		api.GET("/executions/:executionId/history", handlers.GetExecutionHistory)

		// Batch Execution
		api.POST("/state-machines/:stateMachineId/executions/batch", handlers.ExecuteBatch)
		api.POST("/queue/enqueue", handlers.EnqueueExecution)

		// Message/Resume
		api.POST("/executions/:executionId/resume", handlers.ResumeExecution)
		api.POST("/state-machines/:stateMachineId/resume-by-correlation", handlers.ResumeByCorrelation)
		api.GET("/state-machines/:stateMachineId/waiting", handlers.FindWaitingExecutions)
	}

	return router
}

// NewServer creates a new Gin server with state machine routes
func NewServer(config *middleware.Config) *gin.Engine {
	return SetupRouter(config)
}
