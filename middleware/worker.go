package middleware

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/handler"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
)

// WorkerConfig holds configuration for the background worker
type WorkerConfig struct {
	QueueConfig       *queue.Config
	RepositoryManager *repository.Manager
	BaseExecutor      *executor.BaseExecutor
	EnableWorker      bool // Flag to enable/disable worker
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

	if config.BaseExecutor == nil {
		log.Println("Warning: BaseExecutor is nil, worker cannot be created")
		return nil, nil
	}

	// Create execution context adapter
	execAdapter := executor.NewExecutionContextAdapter(config.BaseExecutor)

	queueClient, err := queue.NewClient(config.QueueConfig)

	// Create execution handler with executor
	newExecutionHandlerWithContext := handler.NewExecutionHandlerWithContext(config.RepositoryManager, queueClient, execAdapter)

	// Create queue worker with handler
	queueWorker, err := queue.NewWorker(config.QueueConfig, newExecutionHandlerWithContext)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

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
		log.Fatalf("Failed to start worker: %v", err)
	}

	// Wait for shutdown signal in background
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping worker...")
		w.Stop()
	}()
}
