# Makefile for state-machine-amz-gin
.PHONY: help build test test-unit test-integration test-all clean deps update-deps check-mod \
	fmt vet lint static-check install-lint validate ci \
	docker-up docker-down docker-ps docker-logs \
	build-example test-example \
	asynqmon godoc info version \
	bench pre-release release release-ci

# Variables
PROJECT_NAME := state-machine-amz-gin
MODULE := github.com/hussainpithawala/state-machine-amz-gin
GO := go
GOFLAGS := -v
DOCKER_COMPOSE := docker-compose -f docker-examples/docker-compose.yml
EXAMPLE_DIR := ./examples
ASYNQMON_URL := http://localhost:8080

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

##@ Help

help: ## Show this help message
	@echo '$(BLUE)$(PROJECT_NAME) - Makefile Commands$(NC)'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(YELLOW)<target>$(NC)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

build: ## Build the package
	@echo "$(BLUE)Building package...$(NC)"
	$(GO) build $(GOFLAGS) ./...

build-example: ## Build examples
	@echo "$(BLUE)Building examples...$(NC)"
	$(GO) build $(GOFLAGS) -o bin/example $(EXAMPLE_DIR)

##@ Testing

test: test-all ## Alias for test-all

test-all: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests only (fast, no database)
	@echo "$(BLUE)Running unit tests...$(NC)"
	$(GO) test -short -v ./...

test-integration: docker-up ## Run integration tests with real databases
	@echo "$(BLUE)Running integration tests...$(NC)"
	@sleep 3
	$(GO) test -v ./... -tags=integration || ($(MAKE) docker-down && exit 1)

test-example: build-example ## Run example program
	@echo "$(BLUE)Running example program...$(NC)"
	./bin/example

bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

##@ Code Quality

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)Code formatted successfully$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GO) vet ./...

install-lint: ## Install golangci-lint if not present
	@which golangci-lint > /dev/null || (echo "$(YELLOW)Installing golangci-lint...$(NC)" && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin)

lint: install-lint ## Run linter
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	golangci-lint run ./...

static-check: ## Run staticcheck
	@echo "$(BLUE)Running staticcheck...$(NC)"
	@which staticcheck > /dev/null || (echo "$(YELLOW)Installing staticcheck...$(NC)" && \
		go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

validate: fmt vet lint ## Run all validation checks
	@echo "$(GREEN)All validation checks passed$(NC)"

ci: validate test-all ## Run all CI checks locally
	@echo "$(GREEN)All CI checks passed$(NC)"

##@ Dependencies

deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GO) mod download

update-deps: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy

check-mod: ## Check module dependencies
	@echo "$(BLUE)Checking module dependencies...$(NC)"
	$(GO) mod verify
	$(GO) mod tidy -diff

##@ Docker Infrastructure

docker-up: ## Start PostgreSQL, Redis, and Asynqmon containers for testing
	@echo "$(BLUE)Starting Docker containers...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Containers started successfully$(NC)"
	@echo "$(YELLOW)PostgreSQL available at localhost:5432$(NC)"
	@echo "$(YELLOW)Redis available at localhost:6379$(NC)"
	@echo "$(YELLOW)Asynqmon UI available at $(ASYNQMON_URL)$(NC)"

docker-down: ## Stop test infrastructure containers
	@echo "$(BLUE)Stopping Docker containers...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)Containers stopped successfully$(NC)"

docker-ps: ## Show running containers
	@echo "$(BLUE)Running containers:$(NC)"
	$(DOCKER_COMPOSE) ps

docker-logs: ## Show logs from all containers
	@echo "$(BLUE)Container logs:$(NC)"
	$(DOCKER_COMPOSE) logs -f

asynqmon: ## Open Asynqmon UI in browser
	@echo "$(BLUE)Opening Asynqmon UI...$(NC)"
	@command -v open > /dev/null && open $(ASYNQMON_URL) || \
		(command -v xdg-open > /dev/null && xdg-open $(ASYNQMON_URL)) || \
		echo "$(YELLOW)Please open $(ASYNQMON_URL) in your browser$(NC)"

##@ Utilities

clean: docker-down ## Clean up test artifacts and containers
	@echo "$(BLUE)Cleaning up...$(NC)"
	rm -rf bin/
	rm -rf docker-examples/postgres_14/data/
	$(GO) clean -testcache
	@echo "$(GREEN)Cleanup complete$(NC)"

godoc: ## Start Go documentation server
	@echo "$(BLUE)Starting Go documentation server at http://localhost:6060$(NC)"
	@which godoc > /dev/null || (echo "$(YELLOW)Installing godoc...$(NC)" && \
		go install golang.org/x/tools/cmd/godoc@latest)
	godoc -http=:6060

info: ## Show project information
	@echo "$(BLUE)Project Information:$(NC)"
	@echo "  Name:        $(PROJECT_NAME)"
	@echo "  Module:      $(MODULE)"
	@echo "  Go Version:  $$(go version)"
	@echo "  Git Branch:  $$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'N/A')"
	@echo "  Git Commit:  $$(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"

version: ## Show version information
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "  Go:          $$(go version | awk '{print $$3}')"
	@echo "  Git Commit:  $$(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"
	@echo "  Git Tag:     $$(git describe --tags --abbrev=0 2>/dev/null || echo 'No tags')"
	@echo "  Build Time:  $$(date -u '+%Y-%m-%d %H:%M:%S UTC')"

##@ Release

pre-release: validate test-all ## Prepare for release
	@echo "$(BLUE)Preparing for release...$(NC)"
	@echo "$(GREEN)Pre-release checks passed$(NC)"
	@echo "$(YELLOW)Don't forget to update version tags!$(NC)"

release: pre-release ## Create a new release (interactive)
	@echo "$(BLUE)Creating release...$(NC)"
	@read -p "Enter version tag (e.g., v1.0.0): " VERSION; \
	git tag -a $$VERSION -m "Release $$VERSION"; \
	echo "$(GREEN)Created tag $$VERSION$(NC)"; \
	echo "$(YELLOW)Push with: git push origin $$VERSION$(NC)"

release-ci: pre-release ## Create release for CI (non-interactive)
	@echo "$(BLUE)Creating CI release...$(NC)"
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)ERROR: VERSION variable is required$(NC)"; \
		exit 1; \
	fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "$(GREEN)Release $(VERSION) created and pushed$(NC)"
