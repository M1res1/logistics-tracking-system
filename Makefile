.PHONY: up down migrate test logs ps help

# ── Dev environment ──────────────────────────────────────────────────────────
up: ## Start postgres, redis, kafka (infrastructure only)
	docker compose up -d postgres redis zookeeper kafka
	@echo "⏳  Waiting for postgres to be ready..."
	@docker compose exec postgres sh -c 'until pg_isready -U $$POSTGRES_USER; do sleep 1; done'
	@echo "✅  Infrastructure is up."

down: ## Stop and remove all containers
	docker compose down

up-all: ## Start everything including app services and nginx
	docker compose up -d

# ── Database ─────────────────────────────────────────────────────────────────
migrate: ## Run SQL migration files in order (./migrations/*.sql)
	@echo "▶  Running migrations..."
	@for f in $$(ls migrations/*.sql 2>/dev/null | sort); do \
		echo "  applying $$f"; \
		docker compose exec -T postgres psql \
			-U $$DB_USER -d $$DB_NAME < $$f; \
	done
	@echo "✅  Migrations done."

# ── Testing ──────────────────────────────────────────────────────────────────
test: ## Run all Go tests
	go test -v -race ./...

# ── Helpers ──────────────────────────────────────────────────────────────────
logs: ## Tail logs for all services (pass s=<service> to filter)
	docker compose logs -f $(s)

ps: ## Show running containers
	docker compose ps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
