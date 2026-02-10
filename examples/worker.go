package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
)

// This example demonstrates how to run a standalone worker
// that consumes execution tasks from Redis queue
func runStandaloneWorker() {
	ctx := context.Background()

	// Setup repository manager (PostgreSQL with GORM)
	repoConfig := &repository.Config{
		Strategy:      "gorm-postgres",
		ConnectionURL: "postgres://postgres:postgres@localhost:5432/statemachine_gin?sslmode=disable",
		Options: map[string]interface{}{
			"max_open_conns":    25,
			"max_idle_conns":    5,
			"conn_max_lifetime": 5 * time.Minute,
		},
	}

	repoManager, err := repository.NewPersistenceManager(repoConfig)
	if err != nil {
		log.Fatalf("Failed to create repository manager: %v", err)
	}
	defer repoManager.Close()

	// Initialize database schema
	if err := repoManager.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	log.Println("Repository manager initialized successfully")

	// Setup queue client (Redis)
	queueConfig := &queue.Config{
		RedisClientOpt: &asynq.RedisClientOpt{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		RetryPolicy: &queue.RetryPolicy{
			MaxRetry: 3,
			Timeout:  10 * time.Minute,
		},
	}

	allStateMachines, err := repoManager.ListStateMachines(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list state machines: %v", err)
	}

	for i := 0; i < len(allStateMachines); i++ {
		queueName := allStateMachines[i].ID
		queueConfig.Queues[queueName] = 5
	}

	baseExecutor := executor.NewBaseExecutor()
	log.Println("BaseExecutor initialized with task handler registry")

	// Setup worker configuration
	workerConfig := &middleware.WorkerConfig{
		QueueConfig:       queueConfig,
		RepositoryManager: repoManager,
		BaseExecutor:      baseExecutor,
		EnableWorker:      true,
	}

	// Create worker
	worker, err := middleware.NewWorker(workerConfig)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	if worker == nil {
		log.Fatal("Worker is nil, cannot start")
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker
	if err := worker.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	log.Println("Standalone worker is running. Press Ctrl+C to shutdown gracefully.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping worker...")
	worker.Stop()
	log.Println("Worker stopped successfully")
}
