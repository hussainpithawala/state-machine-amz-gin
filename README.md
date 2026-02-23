# State Machine Amazon States Language Gin Framework

A Gin-based REST API framework for managing and executing state machines using the [state-machine-amz-go](https://github.com/hussainpithawala/state-machine-amz-go) library. This framework provides a complete HTTP interface for state machine orchestration, execution management, and distributed queue processing.

## Features

- 🚀 **RESTful API** for state machine management
- 🔄 **Execution Control** - Start, stop, and monitor executions
- 📊 **State History Tracking** - Complete audit trail of state transitions
- 🔁 **Batch Execution** - Process multiple executions concurrently or via distributed queues
- 💬 **Message-Based Resumption** - Resume paused executions with correlation support
- 🌐 **Distributed Queue Support** - Redis-backed task queue for scalable execution
- 🏥 **Health Monitoring** - Built-in health checks and queue statistics
- 🔌 **Middleware Architecture** - Easy integration with existing Gin applications

## Installation

```bash
go get github.com/hussainpithawala/state-machine-amz-gin
```

## Quick Start

### 1. Create a Simple Server

```go
package main

import (
    "context"
    "log"

    statemachinegin "github.com/hussainpithawala/state-machine-amz-gin"
    "github.com/hussainpithawala/state-machine-amz-gin/middleware"
    "github.com/hussainpithawala/state-machine-amz-go/pkg/repository"
)

func main() {
    ctx := context.Background()

    // Setup repository
    repoConfig := &repository.Config{
        Strategy:      "gorm-postgres",
        ConnectionURL: "postgres://user:password@localhost:5432/statemachine",
    }

    repoManager, err := repository.NewPersistenceManager(repoConfig)
    if err != nil {
        log.Print(err)
    }
    defer repoManager.Close()

    if err := repoManager.Initialize(ctx); err != nil {
        log.Print(err)
    }

    // Setup server
    serverConfig := &middleware.Config{
        RepositoryManager: repoManager,
        BasePath:          "/api/v1",
    }

    router := statemachinegin.NewServer(serverConfig)
    router.Run(":8080")
}
```

### 2. Run the Example

```bash
cd examples
go run main.go
```

## API Endpoints

### State Machine Management

#### Create State Machine
```http
POST /api/v1/state-machines
Content-Type: application/json

{
  "id": "order-processing",
  "name": "Order Processing Workflow",
  "description": "Handles order validation and processing",
  "definition": {
    "Comment": "Order processing state machine",
    "StartAt": "ValidateOrder",
    "States": {
      "ValidateOrder": {
        "Type": "Task",
        "Resource": "arn:aws:lambda:::validate:order",
        "Next": "ProcessPayment"
      },
      "ProcessPayment": {
        "Type": "Task",
        "Resource": "arn:aws:lambda:::process:payment",
        "End": true
      }
    }
  }
}
```

#### Get State Machine
```http
GET /api/v1/state-machines/{stateMachineId}
```

#### List State Machines
```http
GET /api/v1/state-machines?name=order&limit=10
```

### Execution Management

#### Start Execution
```http
POST /api/v1/state-machines/{stateMachineId}/executions
Content-Type: application/json

{
  "name": "order-12345",
  "input": {
    "orderId": "12345",
    "customerId": "CUST-001",
    "amount": 99.99
  }
}
```

**Response:**
```json
{
  "executionId": "order-processing-exec-1234567890",
  "stateMachineId": "order-processing",
  "name": "order-12345",
  "status": "RUNNING",
  "startTime": "2026-01-04T16:00:00Z",
  "input": {...}
}
```

#### Get Execution
```http
GET /api/v1/executions/{executionId}
```

#### List Executions
```http
GET /api/v1/state-machines/{stateMachineId}/executions?status=RUNNING&limit=50&offset=0
```

#### Get Execution History
```http
GET /api/v1/executions/{executionId}/history
```

**Response:**
```json
[
  {
    "id": "exec-id-state-timestamp",
    "executionId": "order-processing-exec-1234567890",
    "stateName": "ValidateOrder",
    "stateType": "Task",
    "status": "SUCCEEDED",
    "input": {...},
    "output": {...},
    "startTime": "2026-01-04T16:00:00Z",
    "endTime": "2026-01-04T16:00:01Z",
    "sequenceNumber": 1
  }
]
```

#### Stop Execution
```http
DELETE /api/v1/executions/{executionId}
```

#### Count Executions
```http
GET /api/v1/state-machines/{stateMachineId}/executions/count?status=SUCCEEDED
```

### Batch Execution

#### Execute Batch
```http
POST /api/v1/state-machines/{stateMachineId}/executions/batch
Content-Type: application/json

{
  "filter": {
    "status": "PENDING",
    "limit": 100
  },
  "namePrefix": "batch-2026-01",
  "concurrency": 10,
  "mode": "distributed",
  "stopOnError": false
}
```

**Response:**
```json
{
  "batchId": "batch-2026-01-1234567890",
  "totalEnqueued": 95,
  "totalFailed": 5,
  "mode": "distributed"
}
```

### Message/Resume Operations

#### Resume Execution
```http
POST /api/v1/executions/{executionId}/resume
Content-Type: application/json

{
  "output": {
    "approved": true,
    "approvedBy": "manager@example.com"
  }
}
```

#### Resume by Correlation
```http
POST /api/v1/state-machines/{stateMachineId}/resume-by-correlation
Content-Type: application/json

{
  "correlationKey": "orderId",
  "correlationValue": "12345",
  "output": {
    "status": "completed"
  }
}
```

**Response:**
```json
{
  "resumedCount": 3,
  "executionIds": ["exec-1", "exec-2", "exec-3"]
}
```

#### Find Waiting Executions
```http
GET /api/v1/state-machines/{stateMachineId}/waiting?correlationKey=orderId&correlationValue=12345
```

### Queue Operations

#### Enqueue Execution
```http
POST /api/v1/queue/enqueue
Content-Type: application/json

{
  "stateMachineId": "order-processing",
  "executionName": "order-98765",
  "input": {
    "orderId": "98765"
  },
  "queue": "critical"
}
```

**Response:**
```json
{
  "taskId": "task-1234567890",
  "queue": "critical",
  "enqueuedAt": "2026-01-04T16:00:00Z"
}
```

#### Get Queue Statistics
```http
GET /api/v1/queue/stats
```

**Response:**
```json
{
  "queues": {
    "critical": {
      "pending": 10,
      "active": 5,
      "failed": 2
    },
    "default": {
      "pending": 50,
      "active": 10,
      "failed": 1
    }
  }
}
```

### Monitoring

#### Health Check
```http
GET /api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "services": {
    "database": "up",
    "queue": "up"
  }
}
```

## Configuration

### Repository Configuration

The framework supports multiple repository backends:

```go
// PostgreSQL with GORM
repoConfig := &repository.Config{
    Strategy:      "gorm-postgres",
    ConnectionURL: "postgres://user:password@localhost:5432/statemachine",
    Options: map[string]interface{}{
        "max_open_conns":    25,
        "max_idle_conns":    5,
        "conn_max_lifetime": 5 * time.Minute,
    },
}

// PostgreSQL with database/sql
repoConfig := &repository.Config{
    Strategy:      "postgres",
    ConnectionURL: "postgres://user:password@localhost:5432/statemachine",
}
```

### Queue Configuration

```go
queueConfig := &queue.Config{
    RedisAddr:     "localhost:6379",
    RedisPassword: "",
    RedisDB:       0,
    Concurrency:   10,
    Queues: map[string]int{
        "critical": 6,  // Priority 6
        "default":  3,  // Priority 3
        "low":      1,  // Priority 1
    },
    RetryPolicy: &queue.RetryPolicy{
        MaxRetry: 3,
        Timeout:  10 * time.Minute,
    },
}
```

### Middleware Configuration

```go
serverConfig := &middleware.Config{
    RepositoryManager: repoManager,
    QueueClient:       queueClient,  // Optional
    BasePath:          "/api/v1",
}
```

## Advanced Usage

### Custom Middleware

Add custom middleware to the router:

```go
router := statemachinegin.NewServer(serverConfig)

// Add authentication middleware
router.Use(authMiddleware())

// Add rate limiting
router.Use(rateLimitMiddleware())

router.Run(":8080")
```

### Integrating with Existing Gin Application

```go
app := gin.Default()

// Your existing routes
app.GET("/", homeHandler)

// Add state machine routes
serverConfig := &middleware.Config{
    RepositoryManager: repoManager,
    BasePath:          "/state-machine",
}

stateMachineRouter := statemachinegin.SetupRouter(serverConfig)

// Mount state machine routes
app.Any("/state-machine/*path", func(c *gin.Context) {
    stateMachineRouter.HandleContext(c)
})

app.Run(":8080")
```

## Error Handling

All endpoints return consistent error responses:

```json
{
  "error": "Error type",
  "message": "Detailed error message",
  "code": 400
}
```

HTTP Status Codes:
- `200 OK` - Successful operation
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service is unhealthy

## Examples

See the [examples](./examples) directory for complete working examples:

- `examples/main.go` - Basic REST API server setup

## Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [state-machine-amz-go](https://github.com/hussainpithawala/state-machine-amz-go) - State machine engine

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
