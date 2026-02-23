# Contributing to state-machine-amz-gin

First off, thank you for considering contributing to state-machine-amz-gin! It's people like you that make this project better for everyone.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Community](#community)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+ (for testing)
- Redis 6+ (for queue testing)
- Git
- Make (optional, but recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/state-machine-amz-gin.git
   cd state-machine-amz-gin
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/hussainpithawala/state-machine-amz-gin.git
   ```

## Development Setup

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make install-tools
```

### 2. Setup Development Environment

```bash
# Copy environment template (if exists)
cp .env.example .env

# Edit .env with your local settings
# Configure database and Redis connections
```

### 3. Setup Development Database

```bash
# Start PostgreSQL and Redis using Docker
docker-compose -f docker-examples/docker-compose.yml up -d postgres redis

# Or use your local installations
# PostgreSQL: createdb statemachine_dev
# Redis: redis-server
```

### 4. Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./handlers/...
```

### 5. Run the Example Server

```bash
# Run the example server
make run-example

# Or manually
cd examples
go run main.go
```

## How to Contribute

### Types of Contributions

We welcome many types of contributions:

- **Bug fixes** - Fix issues found in the codebase
- **New features** - Implement new functionality
- **Documentation** - Improve or add documentation
- **Tests** - Add or improve test coverage
- **Examples** - Add usage examples
- **Performance improvements** - Optimize code
- **Code cleanup** - Refactoring and code quality improvements

### Contribution Workflow

1. **Check existing issues** - Look for existing issues or create a new one
2. **Discuss major changes** - For significant changes, open an issue first to discuss
3. **Create a branch** - Create a feature branch from `main`
4. **Make changes** - Implement your changes with tests
5. **Test locally** - Run tests and linters
6. **Commit** - Follow commit message guidelines
7. **Push** - Push to your fork
8. **Pull request** - Create a pull request to the main repository

## Coding Standards

### Go Style Guide

We follow standard Go conventions:

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) as a guide
- Run `gofmt` and `goimports` before committing
- Follow the project's existing code style

### Code Formatting

```bash
# Format code
make fmt

# Check formatting
make fmt-check

# Run linter
make lint

# Auto-fix linting issues (when possible)
make lint-fix
```

### Code Organization

```
state-machine-amz-gin/
├── handlers/       # HTTP request handlers
├── middleware/     # Gin middleware
├── models/        # Data models
├── examples/      # Usage examples
├── test/          # Integration tests
└── docs/          # Additional documentation
```

### Naming Conventions

- **Files**: `snake_case.go`
- **Packages**: Single word, lowercase
- **Types**: `PascalCase`
- **Functions**: `PascalCase` (exported), `camelCase` (unexported)
- **Variables**: `camelCase`
- **Constants**: `PascalCase` or `SCREAMING_SNAKE_CASE` for package-level constants

### Documentation

- All exported functions, types, and constants must have documentation comments
- Documentation should be clear and concise
- Include examples for complex functionality
- Use complete sentences with proper punctuation

Example:

```go
// ExecutionHandler handles HTTP requests related to state machine executions.
// It provides endpoints for creating, retrieving, updating, and deleting executions.
type ExecutionHandler struct {
    service ExecutionService
    logger  Logger
}

// NewExecutionHandler creates a new ExecutionHandler with the given service and logger.
// The service parameter must not be nil.
//
// Example:
//     handler := NewExecutionHandler(service, logger)
func NewExecutionHandler(service ExecutionService, logger Logger) *ExecutionHandler {
    return &ExecutionHandler{
        service: service,
        logger:  logger,
    }
}
```

### Error Handling

- Always handle errors explicitly
- Use wrapped errors with context: `fmt.Errorf("failed to do X: %w", err)`
- Define custom error types for domain-specific errors
- Don't panic in library code

```go
// Good
exec, err := h.service.GetExecution(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to get execution %s: %w", id, err)
}

// Bad
exec := h.service.GetExecution(ctx, id)  // Ignoring error
```

### Context Usage

- Always pass `context.Context` as the first parameter
- Respect context cancellation
- Add timeout or deadline when appropriate

```go
func (s *Service) CreateExecution(ctx context.Context, req *CreateRequest) (*Execution, error) {
    // Check context first for long operations
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Your implementation
}
```

## Testing Guidelines

### Test Requirements

- All new features must include tests
- Bug fixes should include regression tests
- Maintain or improve code coverage (target: 70%+)
- Write table-driven tests when testing multiple scenarios

### Test Structure

```go
func TestExecutionHandler_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateExecutionRequest
        want    *Execution
        wantErr bool
    }{
        {
            name: "valid execution",
            input: &CreateExecutionRequest{
                Name: "test-exec",
                StateMachineID: "sm-123",
            },
            want: &Execution{
                ID: "exec-123",
                Name: "test-exec",
            },
            wantErr: false,
        },
        {
            name: "missing state machine ID",
            input: &CreateExecutionRequest{
                Name: "test-exec",
            },
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package
go test -v ./handlers/

# Run specific test
go test -v -run TestExecutionHandler_Create ./handlers/

# Run tests with race detector
go test -race ./...

# Run integration tests
make test-integration
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out
```

## Pull Request Process

### Before Submitting

1. **Update documentation** - If you changed APIs, update the documentation
2. **Add tests** - Ensure your changes are tested
3. **Run tests** - All tests must pass
4. **Run linters** - Code must pass linting
5. **Update CHANGELOG** - Add entry to `[Unreleased]` section
6. **Rebase on main** - Ensure your branch is up to date

```bash
# Update your fork
git fetch upstream
git checkout main
git merge upstream/main

# Rebase your feature branch
git checkout feature/my-feature
git rebase main
```

### Pull Request Guidelines

1. **Title** - Use a clear and descriptive title
2. **Description** - Explain what and why, not how
3. **Link issues** - Reference related issues (e.g., "Fixes #123")
4. **Keep it focused** - One feature/fix per PR
5. **Size** - Smaller PRs are easier to review

### PR Template

When creating a PR, include:

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe how you tested your changes

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Code passes linting
- [ ] All tests pass

## Related Issues
Fixes #issue_number
```

### Review Process

1. At least one maintainer approval is required
2. All CI checks must pass
3. Address review feedback promptly
4. Once approved, a maintainer will merge

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Maintenance tasks
- **ci**: CI/CD changes

### Examples

```bash
# Feature
git commit -m "feat(handlers): add execution batch processing endpoint"

# Bug fix
git commit -m "fix(middleware): handle nil pointer in config validation"

# Documentation
git commit -m "docs(readme): add batch execution examples"

# With body
git commit -m "feat(queue): add Redis-based distributed queue

Implements distributed task queue using Asynq for scalable
execution processing across multiple workers.

Closes #42"
```

### Scope

Common scopes in this project:

- `handlers` - HTTP handlers
- `middleware` - Middleware components
- `models` - Data models
- `queue` - Queue-related code
- `examples` - Example code
- `docs` - Documentation
- `ci` - CI/CD pipelines

## Reporting Bugs

### Before Reporting

1. **Check existing issues** - Search for similar issues
2. **Try latest version** - Verify the bug exists in the latest release
3. **Minimal reproduction** - Create a minimal example that reproduces the issue

### Bug Report Template

```markdown
## Description
Clear description of the bug

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Environment
- Go version: 1.25.7.0
- OS: Ubuntu 22.04
- Package version: v1.0.0

## Additional Context
Any other relevant information
```

## Suggesting Features

### Feature Request Template

```markdown
## Problem Statement
Describe the problem this feature would solve

## Proposed Solution
Describe how you envision this feature working

## Alternatives Considered
Other solutions you've considered

## Additional Context
Any other relevant information
```

## Community

### Getting Help

- **GitHub Issues** - For bugs and feature requests
- **GitHub Discussions** - For questions and discussions
- **Stack Overflow** - Tag your question with `state-machine-amz-gin`

### Communication

- Be respectful and constructive
- Focus on the issue, not the person
- Welcome newcomers
- Assume good intentions

## Recognition

Contributors will be recognized in:

- The project's contributors list
- Release notes for significant contributions
- Our README's contributors section

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT License).

---

## Quick Reference

```bash
# Setup
git clone <your-fork>
make install
make install-tools

# Development
make fmt              # Format code
make lint             # Run linter
make test             # Run tests
make test-coverage    # Generate coverage report

# Before committing
make check            # Run all checks
make pre-commit       # Quick pre-commit checks

# Pull request
git push origin feature/my-feature
# Create PR on GitHub
```

---

Thank you for contributing to state-machine-amz-gin! 🎉
