.PHONY: dev build test lint migrate clean

# ── Colours ───────────────────────────────────────────────────
CYAN  := \033[0;36m
RESET := \033[0m

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-18s$(RESET) %s\n", $$1, $$2}'

# ── Development ───────────────────────────────────────────────
dev: ## Start all services in watch mode (requires air, cargo-watch, nodemon)
	@echo "$(CYAN)Starting dev stack...$(RESET)"
	@docker compose -f nexus-infrastructure/docker-compose.yml up -d postgres redis
	@sleep 2
	@$(MAKE) -j4 dev-backend dev-optimizer dev-frontend

dev-backend:
	cd nexus-backend && air

dev-optimizer:
	cd nexus-logistics-optimizer && cargo watch -x run

dev-frontend:
	cd nexus-frontend && npm run dev

# ── Build ─────────────────────────────────────────────────────
build: build-backend build-optimizer build-frontend ## Build all services

build-backend: ## Build Go backend binary
	cd nexus-backend && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/server ./cmd/server

build-optimizer: ## Build Rust optimizer binary
	cd nexus-logistics-optimizer && cargo build --release

build-frontend: ## Build React frontend
	cd nexus-frontend && npm ci && npm run build

build-contracts: ## Compile Solidity contracts
	cd nexus-blockchain && npm ci && npx hardhat compile

# ── Docker Images ─────────────────────────────────────────────
docker-build: ## Build all Docker images
	docker build -t nexus-backend:latest ./nexus-backend
	docker build -t nexus-optimizer:latest ./nexus-logistics-optimizer
	docker build -t nexus-frontend:latest ./nexus-frontend

# ── Testing ───────────────────────────────────────────────────
test: test-backend test-optimizer test-contracts test-e2e ## Run all tests

test-backend: ## Run Go unit + integration tests
	cd nexus-backend && go test -race -coverprofile=coverage.out ./...

test-optimizer: ## Run Rust tests
	cd nexus-logistics-optimizer && cargo test

test-contracts: ## Run Solidity tests
	cd nexus-blockchain && npx hardhat test

test-e2e: ## Run E2E tests
	cd tests && npm ci && npm test

# ── Code Quality ──────────────────────────────────────────────
lint: lint-backend lint-optimizer lint-frontend ## Lint all services

lint-backend:
	cd nexus-backend && golangci-lint run ./...

lint-optimizer:
	cd nexus-logistics-optimizer && cargo clippy -- -D warnings

lint-frontend:
	cd nexus-frontend && npm run lint

# ── Database ──────────────────────────────────────────────────
migrate: ## Run database migrations
	cd nexus-backend && go run ./cmd/migrate

migrate-down: ## Rollback last migration
	cd nexus-backend && go run ./cmd/migrate --down

# ── Blockchain ────────────────────────────────────────────────
deploy-contracts: ## Deploy smart contracts to configured network
	cd nexus-blockchain && npx hardhat run scripts/deploy.js --network $(NETWORK)

# ── Infrastructure ────────────────────────────────────────────
infra-up: ## Start full stack via Docker Compose
	docker compose -f nexus-infrastructure/docker-compose.yml up -d

infra-down: ## Stop full stack
	docker compose -f nexus-infrastructure/docker-compose.yml down

# ── Cleanup ───────────────────────────────────────────────────
clean: ## Remove build artifacts
	rm -rf nexus-backend/bin
	rm -rf nexus-logistics-optimizer/target
	rm -rf nexus-frontend/dist
	rm -rf nexus-blockchain/artifacts nexus-blockchain/cache
