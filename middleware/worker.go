package middleware

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hussainpithawala/state-machine-amz-go/pkg/batch"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/handler"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/types"
	"github.com/redis/go-redis/v9"
)

// WorkerConfig holds configuration for the background worker
type WorkerConfig struct {
	QueueConfig       *queue.Config
	QueueClient       *queue.Client
	RepositoryManager *repository.Manager
	BaseExecutor      *executor.BaseExecutor
	BatchOrchestrator *batch.Orchestrator
	BulkOrchestrator  *batch.Orchestrator
	EnableWorker      bool // Flag to enable/disable worker
	RedisClient       *redis.Client
}

// Worker represents a background worker that consumes from Redis queue
type Worker struct {
	queueWorker *queue.Worker
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorker creates a new background worker instance
func NewWorker(config *WorkerConfig) (*Worker, error) {
	if config == nil {
		return nil, nil
	}

	if !config.EnableWorker {
		log.Println("Worker is disabled in configuration")
		return nil, nil
	}

	if config.QueueConfig == nil {
		log.Println("Warning: QueueConfig is nil, worker cannot be created")
		return nil, nil
	}

	if config.RepositoryManager == nil {
		log.Println("Warning: RepositoryManager is nil, worker cannot be created")
		return nil, nil
	}

	if config.RedisClient == nil {
		log.Println("Warning: RedisClient is nil, worker cannot be created")
		return nil, nil
	}

	if config.BaseExecutor == nil {
		log.Println("Warning: BaseExecutor is nil, worker cannot be created")
		return nil, nil
	}

	if config.QueueClient == nil {
		log.Println("Warning: QueueClient is nil, worker cannot be created")
		return nil, nil
	}

	// Create execution context adapter
	execAdapter := executor.NewExecutionContextAdapter(config.BaseExecutor)

	queueClient, _ := queue.NewClient(config.QueueConfig)

	// Setup micro-batch bulkOrchestrator (optional)
	if config.QueueClient != nil {
		if config.BulkOrchestrator == nil && config.BatchOrchestrator == nil {
			batchOrchestrator, bulkOrchestrator, err := createMiddlewareOrchestrator(context.Background(), config.RepositoryManager, config.QueueClient, config.RedisClient)
			if err != nil {
				log.Printf("Warning: Failed to create bulkOrchestrator: %v (continuing without bulkOrchestrator support)", err)
			} else {
				log.Println("BatchOrchestrator initialized successfully")
			}
			config.BulkOrchestrator = bulkOrchestrator
			config.BatchOrchestrator = batchOrchestrator
		}
	}

	// Create execution handler with executor
	newExecutionHandlerWithContext := handler.NewExecutionHandlerWithContext(
		config.RepositoryManager,
		queueClient,
		execAdapter,
		config.BulkOrchestrator,
	)

	// Create queue worker with handler
	queueWorker, err := queue.NewWorker(config.QueueConfig, newExecutionHandlerWithContext)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	ctx = context.WithValue(ctx, types.ExecutionContextKey, execAdapter)

	return &Worker{
		queueWorker: queueWorker,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start starts the worker in a separate goroutine
func (w *Worker) Start() error {
	if w == nil || w.queueWorker == nil {
		return nil
	}

	log.Println("Starting background worker to consume from Redis queue...")

	// Start worker in goroutine
	go func() {
		if err := w.queueWorker.Run(); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()

	log.Println("Background worker started successfully")
	return nil
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	if w == nil || w.queueWorker == nil {
		return
	}

	log.Println("Stopping background worker...")
	w.cancel()
	w.queueWorker.Shutdown()
	log.Println("Background worker stopped successfully")
}

// StartWithGracefulShutdown starts the worker and sets up graceful shutdown handlers
func (w *Worker) StartWithGracefulShutdown() {
	if w == nil || w.queueWorker == nil {
		return
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker
	if err := w.Start(); err != nil {
		log.Printf("Failed to start worker: %v", err)
	}

	// Wait for shutdown signal in background
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping worker...")
		w.Stop()
	}()
}
