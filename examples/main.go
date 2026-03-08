package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	statemachinegin "github.com/hussainpithawala/state-machine-amz-gin"
	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/batch"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/queue"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/statemachine/persistent"
	"github.com/redis/go-redis/v9"
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
	}

	allStateMachines, err := repoManager.ListStateMachines(ctx, nil)
	if err != nil {
		log.Printf("Failed to list state machines: %v", err)
	}

	for i := 0; i < len(allStateMachines); i++ {
		queueName := allStateMachines[i].ID
		queueConfig.Queues[queueName] = 5
	}

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

	// Setup micro-batch bulkOrchestrator (optional)
	var batchOrchestrator *batch.Orchestrator
	var bulkOrchestrator *batch.Orchestrator
	if queueClient != nil {
		batchOrchestrator, bulkOrchestrator, err = createMiddlewareOrchestrator(ctx, repoManager, queueClient, queueConfig, redisClient)
		if err != nil {
			log.Printf("Warning: Failed to create bulkOrchestrator: %v (continuing without bulkOrchestrator support)", err)
		} else {
			log.Println("BatchOrchestrator initialized successfully")
		}
	}

	// Setup background worker configuration (optional)
	var workerConfig *middleware.WorkerConfig
	if queueClient != nil {
		workerConfig = &middleware.WorkerConfig{
			QueueConfig:       queueConfig,
			RepositoryManager: repoManager,
			BaseExecutor:      baseExecutor,
			BatchOrchestrator: batchOrchestrator,
			BulkOrchestrator:  bulkOrchestrator,
			EnableWorker:      true, // Set to true to enable background worker
			RedisClient:       redisClient,
		}
	}

	// Setup Gin server with state machine middleware
	serverConfig := &middleware.Config{
		RepositoryManager:   repoManager,
		RedisClient:         redisClient,
		QueueClient:         queueClient,
		BaseExecutor:        baseExecutor,
		BatchOrchestrator:   batchOrchestrator,
		BulkOrchestrator:    bulkOrchestrator,
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

func createMiddlewareOrchestrator(
	ctx context.Context,
	repoManager *repository.Manager,
	queueClient *queue.Client,
	queueConfig *queue.Config,
	redisClient *redis.Client,
) (batchOrchestrator *batch.Orchestrator, bulkOrchestrator *batch.Orchestrator, orchError error) {
	if queueConfig == nil || queueConfig.RedisClientOpt == nil {
		return nil, nil, fmt.Errorf("queue redis configuration is required for batchOrchestrator")
	}

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("redis ping failed: %w", err)
	}

	parentSM, err := persistent.New(batch.OrchestratorDefinitionJSON(), true, batch.OrchestratorStateMachineID, repoManager)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create batchOrchestrator parent state machine: %w", err)
	}
	parentSM.SetQueueClient(queueClient)

	smFactory := func(ctx context.Context, smID string, manager *repository.Manager) (batch.StateMachine, error) {
		return persistent.NewFromDefnId(ctx, smID, manager)
	}

	smCreator := func(def []byte, isJSON bool, smID string, manager *repository.Manager) (batch.StateMachine, error) {
		return persistent.New(def, isJSON, smID, manager)
	}

	batchOrchestrator, err = batch.NewOrchestrator(ctx, redisClient, parentSM, smFactory, smCreator)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create batchOrchestrator: %w", err)
	}

	if err := batchOrchestrator.EnsureDefinition(ctx, batch.OrchestratorDefinitionJSON(), batch.OrchestratorStateMachineID); err != nil {
		return nil, nil, fmt.Errorf("failed to register batchOrchestrator definition: %w", err)
	}

	bulkOrchestrator, err = batch.NewOrchestrator(ctx, redisClient, parentSM, smFactory, smCreator)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create bulkOrchestrator: %w", err)
	}
	if err := bulkOrchestrator.EnsureDefinition(ctx, batch.BulkOrchestratorDefinitionJSON(), batch.BulkOrchestratorStateMachineID); err != nil {
		return nil, nil, fmt.Errorf("failed to register bulkOrchestrator definition: %w", err)
	}

	return batchOrchestrator, bulkOrchestrator, nil
}
