package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
)

// Config holds the configuration for the state machine middleware
type Config struct {
	RepositoryManager *repository.Manager
	QueueClient       *queue.Client
	BaseExecutor      *executor.BaseExecutor
	WorkerConfig      *WorkerConfig // Optional: Configuration for background worker
	BasePath          string        // e.g., "/api/v1"
}

// StateMachineMiddleware injects repository manager, queue client, and base executor into gin context
func StateMachineMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.RepositoryManager != nil {
			c.Set("repositoryManager", config.RepositoryManager)
		}
		if config.QueueClient != nil {
			c.Set("queueClient", config.QueueClient)
		}
		if config.BaseExecutor != nil {
			c.Set("baseExecutor", config.BaseExecutor)
		}
		c.Next()
	}
}

// GetRepositoryManager retrieves the repository manager from gin context
func GetRepositoryManager(c *gin.Context) (*repository.Manager, bool) {
	manager, exists := c.Get("repositoryManager")
	if !exists {
		return nil, false
	}
	repoManager, ok := manager.(*repository.Manager)
	return repoManager, ok
}

// GetQueueClient retrieves the queue client from gin context
func GetQueueClient(c *gin.Context) (*queue.Client, bool) {
	client, exists := c.Get("queueClient")
	if !exists {
		return nil, false
	}
	queueClient, ok := client.(*queue.Client)
	return queueClient, ok
}

// GetBaseExecutor retrieves the base executor from gin context
func GetBaseExecutor(c *gin.Context) (*executor.BaseExecutor, bool) {
	exec, exists := c.Get("baseExecutor")
	if !exists {
		return nil, false
	}
	baseExecutor, ok := exec.(*executor.BaseExecutor)
	return baseExecutor, ok
}

// ErrorHandler is a middleware that handles panics and returns proper error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(500, gin.H{
					"error":   "Internal Server Error",
					"message": err,
					"code":    500,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
