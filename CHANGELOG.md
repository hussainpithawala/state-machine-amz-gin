# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

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

[Unreleased]: https://github.com/hussainpithawala/state-machine-amz-gin/compare/v1.0.3...HEAD
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
