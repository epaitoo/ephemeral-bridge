.PHONY: help migrate-up migrate-down migrate-create migrate-force migrate-version migrate-up-prod migrate-down-prod migrate-version-prod db-up db-down

# Load environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database URL for migrations
DATABASE_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5433/$(POSTGRES_DB)?sslmode=disable

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

db-up: ## Start the PostgreSQL database using docker-compose
	docker-compose up -d postgres

db-down: ## Stop the PostgreSQL database
	docker-compose down

migrate-up: ## Run all pending migrations
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback the last migration
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create a new migration file (usage: make migrate-create name=create_users_table)
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir ./migrations -seq $(name)

migrate-up-prod: ## Run all pending migrations on PRODUCTION database
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	@if [ -z "$(PROD_DATABASE_URL)" ]; then \
		echo "Error: PROD_DATABASE_URL is not set. Set it in your .env file."; \
		exit 1; \
	fi
	@echo "⚠ Running migrations on PRODUCTION database..."
	@echo "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
	@sleep 5
	migrate -path ./migrations -database "$(PROD_DATABASE_URL)" up

migrate-down-prod: ## Rollback the last migration on PRODUCTION database
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	@if [ -z "$(PROD_DATABASE_URL)" ]; then \
		echo "Error: PROD_DATABASE_URL is not set. Set it in your .env file."; \
		exit 1; \
	fi
	@echo "⚠ Rolling back migration on PRODUCTION database..."
	@echo "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
	@sleep 5
	migrate -path ./migrations -database "$(PROD_DATABASE_URL)" down 1

migrate-version-prod: ## Show current migration version on PRODUCTION database
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	@if [ -z "$(PROD_DATABASE_URL)" ]; then \
		echo "Error: PROD_DATABASE_URL is not set. Set it in your .env file."; \
		exit 1; \
	fi
	migrate -path ./migrations -database "$(PROD_DATABASE_URL)" version

migrate-force: ## Force set migration version (usage: make migrate-force version=1)
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	migrate -path ./migrations -database "$(DATABASE_URL)" force $(version)

migrate-version: ## Show current migration version
	@if [ -z "$(shell which migrate)" ]; then \
		echo "Error: golang-migrate is not installed. Install with:"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	migrate -path ./migrations -database "$(DATABASE_URL)" version
