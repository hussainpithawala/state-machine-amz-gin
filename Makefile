# Makefile for state-machine-amz-gin

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

# Variables
PROJECT_NAME := state-machine-amz-gin
MODULE := github.com/hussainpithawala/state-machine-amz-gin
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOFMT := gofmt
GOLINT := golangci-lint
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html
DOCKER_COMPOSE := docker-compose -f docker-examples/docker-compose.yml
EXAMPLE_DIR := ./examples
ASYNQMON_URL := http://localhost:8080

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

##@ Development

# Get the golangci-lint binary path from the system
LINT_BIN := $(shell command -v golangci-lint 2> /dev/null)

.PHONY: install
install: ## Install dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod verify
	@echo "$(GREEN)Dependencies installed$(NC)"

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)Dependencies tidied$(NC)"

.PHONY: vendor
vendor: ## Create vendor directory
	@echo "$(BLUE)Creating vendor directory...$(NC)"
	$(GO) mod vendor
	@echo "$(GREEN)Vendor directory created$(NC)"

.PHONY: build
build: ## Build the package
	@echo "$(BLUE)Building package...$(NC)"
	$(GO) build -v ./...
	@echo "$(GREEN)Build successful$(NC)"

.PHONY: build-example
build-example: ## Build example server
	@echo "$(BLUE)Building example server...$(NC)"
	$(GO) build -v -o bin/example $(EXAMPLE_DIR)
	@echo "$(GREEN)Example built: bin/example$(NC)"

##@ Code Quality

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) -w -s .
	$(GO) fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "$(BLUE)Checking code format...$(NC)"
	@test -z "$$($(GOFMT) -l .)" || (echo "$(RED)Code not formatted. Run 'make fmt'$(NC)"; exit 1)
	@echo "$(GREEN)Code formatting OK$(NC)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)Vet passed$(NC)"

install-lint:
ifndef LINT_BIN
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/local/bin latest
else
	@golangci-lint version
	@echo "golangci-lint is already installed."
endif

.PHONY: lint
lint: install-lint ## Run linter
	@echo "${GREEN}Running linter...${RESET}"
	@golangci-lint run --timeout=5m --config=.golangci.yml

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "$(BLUE)Running golangci-lint with auto-fix...$(NC)"
	$(GOLINT) run --fix ./...
	@echo "$(GREEN)Lint fixes applied$(NC)"

##@ Testing

.PHONY: test
test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	$(GOTEST) -v -race ./...

.PHONY: test-short
test-short: ## Run short tests (no integration)
	@echo "$(BLUE)Running short tests...$(NC)"
	$(GOTEST) -v -short ./...

.PHONY: test-integration
test-integration: docker-up ## Run integration tests with Docker infrastructure
	@echo "$(BLUE)Running integration tests...$(NC)"
	@sleep 3
	$(GOTEST) -v ./... -tags=integration || ($(MAKE) docker-down && exit 1)

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(NC)"

.PHONY: test-coverage-func
test-coverage-func: ## Show coverage by function
	@echo "$(BLUE)Generating coverage by function...$(NC)"
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

##@ Quality Checks

.PHONY: check
check: fmt-check vet lint test ## Run all quality checks

.PHONY: ci
ci: install check test-coverage ## Run CI pipeline locally

##@ Security

.PHONY: sec
sec: ## Run security checks with gosec
	@echo "$(BLUE)Running security checks...$(NC)"
	@which gosec > /dev/null || (echo "$(YELLOW)Installing gosec...$(NC)" && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -quiet ./...
	@echo "$(GREEN)Security check passed$(NC)"

.PHONY: vuln
vuln: ## Check for known vulnerabilities
	@echo "$(BLUE)Checking for vulnerabilities...$(NC)"
	@which govulncheck > /dev/null || (echo "$(YELLOW)Installing govulncheck...$(NC)" && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...
	@echo "$(GREEN)Vulnerability check passed$(NC)"

##@ Docker Infrastructure

.PHONY: docker-up
docker-up: ## Start PostgreSQL, Redis, and Asynqmon containers for testing
	@echo "$(BLUE)Starting Docker containers...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Containers started successfully$(NC)"
	@echo "$(YELLOW)PostgreSQL available at localhost:5432$(NC)"
	@echo "$(YELLOW)Redis available at localhost:6379$(NC)"
	@echo "$(YELLOW)Asynqmon UI available at $(ASYNQMON_URL)$(NC)"

.PHONY: docker-down
docker-down: ## Stop test infrastructure containers
	@echo "$(BLUE)Stopping Docker containers...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)Containers stopped successfully$(NC)"

.PHONY: docker-ps
docker-ps: ## Show running containers
	@echo "$(BLUE)Running containers:$(NC)"
	$(DOCKER_COMPOSE) ps

.PHONY: docker-logs
docker-logs: ## Show logs from all containers
	@echo "$(BLUE)Container logs:$(NC)"
	$(DOCKER_COMPOSE) logs -f

.PHONY: asynqmon
asynqmon: ## Open Asynqmon UI in browser
	@echo "$(BLUE)Opening Asynqmon UI...$(NC)"
	@command -v open > /dev/null && open $(ASYNQMON_URL) || \
		(command -v xdg-open > /dev/null && xdg-open $(ASYNQMON_URL)) || \
		echo "$(YELLOW)Please open $(ASYNQMON_URL) in your browser$(NC)"

##@ Documentation

.PHONY: docs
docs: ## Serve documentation locally at http://localhost:6060
	@echo "$(BLUE)Starting documentation server...$(NC)"
	@echo "$(GREEN)Visit: http://localhost:6060/pkg/$(MODULE)/$(NC)"
	@which godoc > /dev/null || (echo "$(YELLOW)Installing godoc...$(NC)" && go install golang.org/x/tools/cmd/godoc@latest)
	godoc -http=:6060

##@ Release

.PHONY: release-check
release-check: ## Check if ready for release
	@echo "$(BLUE)Checking release readiness...$(NC)"
	@test -f CHANGELOG.md || (echo "$(RED)❌ CHANGELOG.md not found$(NC)"; exit 1)
	@test -f LICENSE || (echo "$(RED)❌ LICENSE not found$(NC)"; exit 1)
	@test -f CONTRIBUTING.md || (echo "$(YELLOW)⚠️  CONTRIBUTING.md recommended$(NC)")
	@$(MAKE) check
	@echo "$(GREEN)✅ Release checks passed$(NC)"

.PHONY: tag
tag: ## Create git tag (Usage: make tag VERSION=v1.0.0)
	@test -n "$(VERSION)" || (echo "$(RED)VERSION required. Usage: make tag VERSION=v1.0.0$(NC)"; exit 1)
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "$(GREEN)✅ Created tag $(VERSION)$(NC)"
	@echo "$(YELLOW)Push with: git push origin $(VERSION)$(NC)"

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts and caches
	@echo "$(BLUE)Cleaning up...$(NC)"
	rm -rf bin/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	$(GO) clean -cache -testcache
	@echo "$(GREEN)Cleanup complete$(NC)"

.PHONY: clean-all
clean-all: clean docker-down ## Deep clean including Docker and mod cache
	@echo "$(BLUE)Deep cleaning...$(NC)"
	rm -rf docker-examples/postgres_14/data/
	$(GO) clean -modcache
	@echo "$(GREEN)Deep cleanup complete$(NC)"

##@ Tools

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/godoc@latest
	@echo "$(GREEN)✅ Tools installed$(NC)"

##@ Dependencies

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

.PHONY: deps-check
deps-check: ## Check for outdated dependencies
	@echo "$(BLUE)Checking for outdated dependencies...$(NC)"
	$(GO) list -u -m all

##@ Utilities

.PHONY: info
info: ## Show project information
	@echo "$(BLUE)Project Information:$(NC)"
	@echo "  Name:        $(PROJECT_NAME)"
	@echo "  Module:      $(MODULE)"
	@echo "  Go Version:  $$(go version)"
	@echo "  Git Branch:  $$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'N/A')"
	@echo "  Git Commit:  $$(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"

.PHONY: version
version: ## Show version information
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "  Go:          $$(go version | awk '{print $$3}')"
	@echo "  Git Commit:  $$(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"
	@echo "  Git Tag:     $$(git describe --tags --abbrev=0 2>/dev/null || echo 'No tags')"
	@echo "  Build Time:  $$(date -u '+%Y-%m-%d %H:%M:%S UTC')"

##@ Pre-commit/push Hooks

.PHONY: pre-commit
pre-commit: fmt vet lint test-short ## Quick pre-commit checks
	@echo "$(GREEN)✅ Pre-commit checks passed$(NC)"

.PHONY: pre-push
pre-push: check test-coverage ## Pre-push checks
	@echo "$(GREEN)✅ Pre-push checks passed$(NC)"
