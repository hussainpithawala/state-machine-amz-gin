package main

import (
	"context"
	"fmt"
	"log"
	"time"

	statemachinegin "github.com/hussainpithawala/state-machine-amz-gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
)

func main() {
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
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		Concurrency:   10,
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

	queueClient, err := queue.NewClient(queueConfig)
	if err != nil {
		log.Printf("Warning: Failed to create queue client: %v (continuing without queue support)", err)
		queueClient = nil
	} else {
		defer queueClient.Close()
		log.Println("Queue client initialized successfully")
	}

	// Setup Gin server with state machine middleware
	serverConfig := &middleware.Config{
		RepositoryManager: repoManager,
		QueueClient:       queueClient,
		BasePath:          "/state-machines/api/v1",
	}

	router := statemachinegin.NewServer(serverConfig)

	// Start server
	port := 9090
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Starting state machine REST API server on %s", addr)
	log.Printf("API Documentation available at http://localhost%s/health", addr)
	log.Printf("Example endpoints:")
	log.Printf("  - POST   http://localhost%s/api/v1/state-machines", addr)
	log.Printf("  - GET    http://localhost%s/api/v1/state-machines", addr)
	log.Printf("  - POST   http://localhost%s/api/v1/state-machines/:id/executions", addr)
	log.Printf("  - GET    http://localhost%s/api/v1/executions/:id", addr)
	log.Printf("  - GET    http://localhost%s/api/v1/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
