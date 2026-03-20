# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

### Changed
- **Dependency Update** - Updated `state-machine-amz-go` to v1.2.13

### Deprecated
- Nothing yet

### Removed
- Nothing yet

### Fixed
- Nothing yet

### Security
- Nothing yet

## [1.0.9] - 2026-03-18

### Added
- **ResumeOrchestrator API** - New endpoint to resume stuck orchestrators handling exceptional conditions
  - `POST /orchestrator/resume` - Resume orchestrators stuck at `WaitForMicroBatchCompletion`/`WaitForBulkMicroBatchCompletion`
  - Pushes orchestrators to handle exceptional conditions when micro-batch completion signals are lost or delayed
- **Redis Signaling Control API** - Human-in-the-loop pause/resume cycle for batch streaming executions
  - `POST /batch/resume/signal` - Signal a paused batch to resume (JSON body)
  - `POST /batch/:batchId/resume/signal` - Signal resume (path parameter + query string)
  - `POST /batch/resume/revoke` - Revoke unconsumed resume signal (JSON body)
  - `POST /batch/:batchId/resume/revoke` - Revoke signal (path parameter)
  - `POST /batch/resume/check` - Check and atomically consume resume signal using Redis GETDEL
  - `GET /batch/:batchId/resume/check` - Check resume signal (path parameter)
- **New Request Models**
  - `ResumeOrchestratorRequest` - Request to resume stuck orchestrator with batch/micro-batch IDs
  - `SignalResumeRequest` - Signal resume with operator and optional notes
  - `RevokeResumeRequest` - Revoke resume signal for a batch
  - `CheckResumeRequest` - Check for resume signal
- **New Response Models**
  - `ResumeOrchestratorResponse` - Response with resume status and message
  - `SignalResumeResponse` - Response with signal confirmation and timestamp
  - `RevokeResumeResponse` - Response with revoke confirmation
  - `CheckResumeResponse` - Response with `shouldResume` flag and signal metadata
- **New Handler File**
  - `handlers/resume.go` - Dedicated handler for Redis-based batch streaming control
- **Postman Collection v3** - New Postman collection with all orchestration and signaling endpoints

### Changed
- **Dependency Update** - Updated `state-machine-amz-go` to v1.2.12 for ResumeController support
- **Docker Compose** - Re-enabled `state-machine-amz-portal` service in `docker-examples/docker-compose.yml`

### Technical Details
- **Atomic Signal Consumption** - Uses Redis GETDEL to ensure resume signals are consumed exactly once
- **Dual API Style** - Both JSON body and path parameter endpoints for operational flexibility
- **Operator Tracking** - Resume signals include operator ID and timestamp for audit trails

## [1.0.8] - 2026-03-12

### Changed
- **Simplified orchestrator initialization** - Moved orchestrator creation from `examples/main.go` to `middleware/worker.go`, centralizing the setup logic within the worker initialization
- **Bulk execution handler fix** - Fixed bulk execution to use `config.BulkOrchestrator` instead of `config.BatchOrchestrator` in `NewExecutionHandlerWithContext` call
- **Improved error handling** - Better error messages and early returns in bulk/batch execution handlers with Redis and Queue client retrieval moved earlier in handlers
- **Queue client validation** - Added nil check for `QueueClient` in worker creation to prevent initialization failures

### Removed
- **Unused imports** - Removed `batch` and `persistent` package imports from `examples/main.go` that are no longer needed at the application level
- **Unused dependency** - Removed `github.com/stretchr/objx v0.5.2` from `go.mod`

### Added
- **Docker portal service** - Enabled `state-machine-amz-portal` service in `docker-compose.yml` with updated image tag `latest`

## [1.0.7] - 2026-03-09

### Added
- **Bulk Execution API** - New orchestration-based bulk execution with micro-batch processing support
  - `POST /state-machines/:stateMachineId/executions/bulk` - Execute bulk operations with JSON input
  - `POST /state-machines/:stateMachineId/executions/bulk-form` - Execute bulk operations with file upload
  - `GET /bulk/:orchestratorId/status` - Get bulk execution status with progress and metrics
  - `POST /bulk/:orchestratorId/pause` - Pause a running bulk execution
  - `POST /bulk/:orchestratorId/resume` - Resume a paused bulk execution
  - `DELETE /bulk/:orchestratorId` - Cancel a bulk execution
  - `GET /bulk` - List all bulk executions
- **Batch Management API** - Enhanced batch execution control
  - `GET /batch/:batchId/status` - Get batch execution status
  - `POST /batch/:batchId/pause` - Pause a running batch
  - `POST /batch/:batchId/resume` - Resume a paused batch
  - `DELETE /batch/:batchId` - Cancel a batch execution
  - `GET /batch` - List all batch executions
- **Micro-batch Orchestration** - Support for processing large input sets in micro-batches
  - Configurable micro-batch size via `microBatchSize` parameter
  - Automatic failure rate monitoring with `pauseThreshold` for auto-pause
  - Configurable resume strategies: `manual`, `automatic`, `timeout`
- **New Response Models** - Comprehensive response types for bulk operations
  - `BulkExecutionResponse` - Initial bulk execution response with orchestrator ID
  - `BulkStatusResponse` - Detailed status with progress and metrics
  - `BulkProgress` - Batch and execution progress tracking
  - `BulkMetrics` - Success/failure rates and performance metrics
  - `BulkActionResponse` - Standardized response for pause/resume/cancel actions
- **Enhanced Request Models** - New request parameters for bulk/batch operations
  - `ExecuteBulkRequest` - Support for `doMicroBatch`, `microBatchSize`, `pauseThreshold`, `resumeStrategy`, `timeoutSeconds`
  - `ExecuteBatchRequest` - Added `doMicroBatch` and `microBatchSize` parameters
- **New Examples**
  - `examples/main.go` - Complete example with bulk orchestrator setup and Redis integration
  - `examples/sm-mac-A-generator/main.go` - Order data generator for testing bulk operations
- **Postman Collection v2** - Updated Postman collection with Bulk Operations endpoints
- **Redis Client Integration** - Direct Redis client support for orchestrator functionality

### Changed
- **Router Updates** - Changed default base path from `/api/v1` to `/api` for cleaner URLs
- **Middleware Configuration** - Enhanced `middleware.Config` with:
  - `RedisClient` - Direct Redis client for orchestrator operations
  - `BatchOrchestrator` - Micro-batch orchestrator for distributed processing
  - `BulkOrchestrator` - Bulk orchestrator for large-scale executions
  - `BaseExecutor` - Base executor for task handler registry
- **Worker Configuration** - Enhanced `WorkerConfig` with orchestrator support
- **State Machine Middleware** - Updated to support optional orchestrator hooks
- **Updated Dependencies** - `state-machine-amz-go` with batch orchestration support

### Changed
- Nothing yet

### Deprecated
- Nothing yet

### Removed
- Nothing yet

### Fixed
- Nothing yet

### Security
- Nothing yet

## [1.0.6] - 2026-02-24

### Changed
- Updated `state-machine-amz-go` to v1.2.10 with local replacement
- Commented out repository and queue client close logic in examples

## [1.0.5] - 2026-02-24

### Changed
- Updated `state-machine-amz-go` to v1.2.9

## [1.0.4] - 2026-02-23

### Added
- Support for `applyUnique` flag in batch handler
- New dependency: `github.com/zeebo/xxh3 v1.0.2`

### Changed
- Updated `state-machine-amz-go` to v1.2.8
- Refactored execution filter logic
- Cleaned up unused modules in go.mod

## [1.0.3] - 2026-02-16

### Changed
- Updated dependencies and cleaned up go.mod

### Removed
- Unused module replacements and old dependency versions

## [1.0.2] - 2026-02-10

### Added
- Examples for state machines A and B with sample order data

### Fixed
- Batch handler filter mapping logic

## [1.0.1] - 2026-02-10

### Changed
- Updated dependencies
- Replaced Redis configuration for queue setup with asynq

## [1.0.0] - 2026-02-09

### Added
- Initial stable release of state-machine-amz-gin
- RESTful API for state machine management
  - Create, read, update, delete state machines
  - List state machines with filtering and pagination
- Execution control endpoints
  - Start new executions
  - Stop running executions
  - Get execution details
  - List executions with filtering
  - Count executions by status
- State history tracking
  - Complete audit trail of state transitions
  - Retrieve execution history with sequencing
  - Track state inputs, outputs, and timestamps
- Batch execution support
  - Execute multiple state machines concurrently
  - Distributed queue-based batch processing
  - Configurable concurrency and error handling
- Message-based resumption system
  - Resume paused executions with new data
  - Correlation-based resumption (resume by business key)
  - Find waiting executions by correlation
- Distributed queue support via Redis
  - Async task execution with Asynq
  - Multiple queue priorities
  - Configurable retry policies
  - Queue statistics and monitoring
- Health monitoring endpoints
  - Service health checks
  - Database connectivity checks
  - Queue system health status
- Complete middleware architecture
  - Easy integration with existing Gin applications
  - Configurable base paths
  - Custom middleware support
- Comprehensive documentation
  - API documentation with examples
  - OpenAPI 3.0 specification
  - Postman collection
  - Docker examples
  - Usage examples
- PostgreSQL persistence layer
  - Support for GORM and database/sql
  - Connection pooling and optimization
  - Automatic schema initialization

### Changed
- N/A (initial release)

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- Input validation on all endpoints
- SQL injection prevention via parameterized queries
- Proper error handling without sensitive data exposure

---

## How to Use This Changelog

### For Maintainers

When making changes:

1. Add entries to the `[Unreleased]` section
2. Use the appropriate category (Added, Changed, Deprecated, Removed, Fixed, Security)
3. Write clear, user-focused descriptions
4. Include references to issues/PRs when applicable

When releasing:

1. Move items from `[Unreleased]` to a new version section
2. Set the release date
3. Update comparison links at the bottom
4. Commit with message: "chore: update CHANGELOG for vX.Y.Z"

### Categories Explained

- **Added** - New features or functionality
- **Changed** - Changes to existing functionality
- **Deprecated** - Features that will be removed in future versions
- **Removed** - Features that have been removed
- **Fixed** - Bug fixes
- **Security** - Security-related changes

### Version Format Examples

```markdown
## [1.2.0] - 2026-03-15

### Added
- New endpoint for bulk state machine updates (#123)
- Support for custom state transition validators
- Metrics collection via Prometheus

### Changed
- Improved error messages for validation failures
- Updated state-machine-amz-go dependency to v1.2.0

### Fixed
- Race condition in concurrent execution handling (#145)
- Memory leak in long-running executions
```

### Breaking Changes

For changes that break backward compatibility, use this format:

```markdown
## [2.0.0] - 2026-06-01

### Changed
- **BREAKING**: Renamed `CreateStateMachine` to `NewStateMachine` for consistency
- **BREAKING**: Changed execution status enum values from uppercase to TitleCase
- **BREAKING**: Removed deprecated `ExecuteSync` method

### Migration Guide

#### Renamed Methods
**Before:**
```go
sm, err := CreateStateMachine(config)
```

**After:**
```go
sm, err := NewStateMachine(config)
```

#### Status Values
**Before:**
```go
if execution.Status == "RUNNING" { ... }
```

**After:**
```go
if execution.Status == "Running" { ... }
```
```

---

## Links

[Unreleased]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.8...HEAD
[1.0.8]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.7...v1.0.8
[1.0.7]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/hussainpithawala/state-machine-amz-gin/releases/tag/v1.0.0

---

## Template for Future Releases

Use this template when creating a new release section:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- Feature A with description (#PR)
- Feature B with description (#PR)

### Changed
- Change A with description (#PR)
- Change B with description (#PR)

### Deprecated
- Feature X (will be removed in vN.0.0)

### Removed
- Feature Y (deprecated since vN.0.0)

### Fixed
- Bug A description (#issue)
- Bug B description (#issue)

### Security
- Security fix description (#issue)

[X.Y.Z]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/vX.Y.Z-1...vX.Y.Z
```
