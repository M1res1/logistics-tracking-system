.PHONY: up down up-all migrate test test-auth run-auth run-order run-deliveries run-restaurants run-payments logs ps help

# ── Dev environment ──────────────────────────────────────────────────────────
up: ## Start postgres, redis, kafka (infrastructure only)
	docker compose up -d postgres redis zookeeper kafka
	@echo "Waiting for postgres..."
	@docker compose exec postgres sh -c 'until pg_isready -U $$POSTGRES_USER; do sleep 1; done'
	@echo "Infrastructure is up."

down: ## Stop and remove all containers
	docker compose down

up-all: ## Start everything including app services and nginx
	docker compose up -d

# ── Database ─────────────────────────────────────────────────────────────────
migrate: ## Run SQL migration files in order (backend/migrations/*.sql)
	@echo "Running migrations..."
	@for f in $$(ls backend/migrations/*.sql 2>/dev/null | sort); do \
		echo "  applying $$f"; \
		docker compose exec -T postgres psql \
			-U $${DB_USER:-postgres} -d $${DB_NAME:-logistics} < $$f; \
	done
	@echo "Migrations done."

# ── Testing ──────────────────────────────────────────────────────────────────
test: ## Run all tests for main module
	cd backend && go test -v -race ./...

test-auth: ## Run auth-go module tests
	cd backend/services/auth-go && go test -v -race ./...

test-all: test test-auth ## Run all tests across all modules

# ── Run individual services locally ─────────────────────────────────────────
run-auth: ## Run auth service
	cd backend/services/auth-go && go run ./cmd

run-order: ## Run order service
	cd backend && go run ./services/order

run-deliveries: ## Run delivery service
	cd backend && go run ./services/delivery

run-restaurants: ## Run restaurant service
	cd backend && go run ./services/restaurant

run-payments: ## Run payment service
	cd backend && go run ./services/payment

# ── Helpers ──────────────────────────────────────────────────────────────────
logs: ## Tail logs (pass s=<service> to filter)
	docker compose logs -f $(s)

ps: ## Show running containers
	docker compose ps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
