package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	statemachinegin "github.com/hussainpithawala/state-machine-amz-gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/recovery"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// Recovery scanner configuration
	const (
		scanInterval        = 5 * time.Second  // How often to scan for orphaned executions
		orphanedThreshold   = 10 * time.Second // Time after which RUNNING execution is considered orphaned
		maxRecoveryAttempts = 3                // Maximum recovery attempts
	)

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
		log.Printf("Failed to create repository manager: %v", err)
	}
	// defer func(repoManager *repository.Manager) {
	// 	err := repoManager.Close()
	// 	if err != nil {
	// 		log.Printf("Warning: Failed to close repository manager: %v", err)
	// 	}
	// }(repoManager)

	// Initialize database schema
	if err := repoManager.Initialize(ctx); err != nil {
		log.Printf("Failed to initialize repository: %v", err)
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
		GroupAggregation: &queue.GroupAggregationConfig{
			Enabled:          true,
			GroupMaxSize:     10_000,
			GroupMaxDelay:    2 * time.Second,
			GroupGracePeriod: 1 * time.Second,
		},
	}

	allStateMachines, err := repoManager.ListStateMachines(ctx, nil)
	if err != nil {
		log.Printf("Failed to list state machines: %v", err)
	}

	// Create persistent state machines and start recovery scanner for each
	var persistentStateMachines []*persistent.StateMachine
	for i := 0; i < len(allStateMachines); i++ {
		queueName := allStateMachines[i].ID
		queueConfig.Queues[queueName] = 5
		queueConfig.Concurrency = 10
		queueConfig.RetryPolicy = &queue.RetryPolicy{}

		// Create persistent state machine with recovery support
		sm := allStateMachines[i]
		psm, err := persistent.NewFromDefnId(ctx, sm.ID, repoManager)
		if err != nil {
			log.Printf("Warning: failed to create persistent state machine for %s: %v", sm.ID, err)
			continue
		}
		persistentStateMachines = append(persistentStateMachines, psm)

		// Configure and start recovery scanner for this state machine
		recoveryConfig := &recovery.RecoveryConfig{
			Enabled:                    true,
			ScanInterval:               scanInterval,
			OrphanedThreshold:          orphanedThreshold,
			DefaultRecoveryStrategy:    recovery.StrategyRetry,
			DefaultMaxRecoveryAttempts: maxRecoveryAttempts,
			StateMachineID:             sm.ID,
		}

		if err := psm.StartRecoveryScanner(recoveryConfig); err != nil {
			log.Printf("Failed to start recovery scanner for %s: %v", sm.ID, err)
		} else {
			log.Printf("Recovery scanner started for state machine: %s", sm.ID)
		}
	}

	// Defer stopping all recovery scanners
	defer func() {
		for _, psm := range persistentStateMachines {
			if err := psm.StopRecoveryScanner(); err != nil {
				log.Printf("Warning: failed to stop recovery scanner for %s: %v", psm.GetID(), err)
			}
		}
	}()

	queueClient, err := queue.NewClient(queueConfig)
	if err != nil {
		log.Printf("Warning: Failed to create queue client: %v (continuing without queue support)", err)
		queueClient = nil
	} else {
		/*		defer func(queueClient *queue.Client) {
					err := queueClient.Close()
					if err != nil {
						log.Printf("Warning: Failed to close queue client: %v", err)
					}
				}(queueClient)
		*/log.Println("Queue client initialized successfully")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         queueConfig.RedisClientOpt.Addr,
		Password:     queueConfig.RedisClientOpt.Password,
		DB:           queueConfig.RedisClientOpt.DB,
		TLSConfig:    queueConfig.RedisClientOpt.TLSConfig,
		ReadTimeout:  queueConfig.RedisClientOpt.ReadTimeout,
		WriteTimeout: queueConfig.RedisClientOpt.WriteTimeout,
		PoolSize:     queueConfig.RedisClientOpt.PoolSize,
	})

	// Create BaseExecutor with StateRegistry for all task handlers
	baseExecutor := executor.NewBaseExecutor()
	RegisterGlobalFunctions(baseExecutor)
	log.Println("BaseExecutor initialized with task handler registry")

	// Setup background worker configuration (optional)
	var workerConfig *middleware.WorkerConfig
	if queueClient != nil {
		workerConfig = &middleware.WorkerConfig{
			QueueConfig:       queueConfig,
			RepositoryManager: repoManager,
			BaseExecutor:      baseExecutor,
			EnableWorker:      true, // Set to true to enable background worker
			RedisClient:       redisClient,
			QueueClient:       queueClient,
		}
	}

	// Setup Gin server with state machine middleware
	serverConfig := &middleware.Config{
		RepositoryManager:   repoManager,
		RedisClient:         redisClient,
		QueueClient:         queueClient,
		BaseExecutor:        baseExecutor,
		WorkerConfig:        workerConfig,
		BasePath:            "/state-machines/api/v1",
		TransformerRegistry: RegisterTransformerFunctions(),
	}

	// Create and start background worker if configured
	worker, err := middleware.NewWorker(workerConfig)
	if err != nil {
		log.Printf("Failed to create worker: %v", err)
	}
	if worker != nil {
		defer worker.Stop()
		if err := worker.Start(); err != nil {
			log.Printf("Failed to start worker: %v", err)
		}
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
		log.Printf("Failed to start server: %v", err)
	}
}
