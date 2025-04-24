# Makefile for Inbox Whisperer

# Install all dependencies and fonts
install:
	npm install --prefix web

# Start the dev server
start:
	npm run dev --prefix web -- --host

# Full setup (install + start)
setup: install start
#
# Usage:
#   make <target>
#
# Key targets:
#   ci                  Run all backend and frontend lint, typecheck, and tests (canonical local CI)
#   ui-install          Install frontend dependencies
#   ui-dev              Start React dev server
#   ui-build            Build React app for production
#   ui-lint             Lint frontend code
#   ui-test             Run frontend tests
#   ui-typecheck        Typecheck frontend code
#   ui-coverage         Generate frontend coverage report
#   ui-generate-api-client  Generate TypeScript API client from OpenAPI spec
#   lint                Lint backend Go code
#   vet                 Run go vet static analysis
#   staticcheck         Run staticcheck static analysis
#   lint-strict         Run all Go lint/static checks (lint, vet, staticcheck)
#   test                Run backend Go tests
#   dev-up              Start DB and apply migrations (local/dev)
#   migrate-create      Create a new DB migration
#   tidy                Run go mod tidy
#   clean               Remove temp/test output files
#   help                Show this help message
#
# For more info, see README.md and migrations/README.md


help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_-]+:|^# ' Makefile | \
	  sed -E 's/^([a-zA-Z0-9_-]+):.*/\1/;s/^# //' | \
	  awk 'BEGIN{t=""} /^[a-zA-Z0-9_-]+$/ {if(t!=""){print t}; t=$0; next} {t=t" "$0} END{print t}' | \
	  column -t -s ':'

clean:
	rm -f test-output.txt

.PHONY: db-up db-init

# Makefile for Inbox Whisperer DB management
# -----------------------------------------
# LOCAL DEVELOPMENT: Use only the psql-migrate-up target for applying migrations.
# This works reliably in devcontainers, VSCode, and all local setups.
#
# CI/CD/Production: Use the docker-migrate-* targets (ensure migration files are present on the host).

# Start the database container
.PHONY: db-up
db-up:
	docker-compose up -d db

# Bring up the entire dev stack (frontend, backend, db) and apply all migrations (IDEMPOTENT)
.PHONY: dev-up
dev-up:
	docker-compose up -d
	for f in $(sort $(wildcard migrations/*_*.up.sql)); do \
	  echo "Applying $$f"; \
	  cat $$f | docker-compose exec -T db psql -U inbox -d inboxwhisperer 2>&1 | grep -v 'already exists' || true; \
	done

# Bring down all dev containers and remove volumes
.PHONY: dev-down
dev-down:
	docker-compose down -v

# Apply all migrations using psql (RECOMMENDED for local/dev)
.PHONY: psql-migrate-up
psql-migrate-up:
	docker-compose up -d db
	for f in $(sort $(wildcard migrations/*_*.up.sql)); do \
	  echo "Applying $$f"; \
	  cat $$f | docker-compose exec -T db psql -U inbox -d inboxwhisperer; \
	done

# Create a new migration (creates up/down SQL files)
.PHONY: migrate-create
migrate-create:
	@read -p "Migration name: " name; migrate create -dir ./migrations -ext sql $$name

# Run all backend and frontend lint, typecheck, and tests (canonical local CI)
.PHONY: ci
ci: lint vet staticcheck tidy test ui-install ui-lint ui-typecheck ui-test ui-build

.PHONY: ui-install ui-lint ui-typecheck ui-test ui-build
ui-install:
	npm install --prefix web

ui-lint:
	npm run lint --prefix web

ui-typecheck:
	npm run typecheck --prefix web

ui-test:
	npm test --prefix web

ui-build:
	npm run build --prefix web

# Lint the codebase using golangci-lint (idempotent)
.PHONY: lint vet staticcheck lint-strict
lint:
	golangci-lint run ./...

# Run go vet static analysis
vet:
	go vet ./...

# Run staticcheck static analysis (requires staticcheck to be installed)
staticcheck:
	staticcheck ./...

# Run all Go lint/static checks
lint-strict: lint vet staticcheck

# Run all tests (idempotent)
.PHONY: test
test:
	go test -v ./...

# Regenerate gomock mocks (idempotent)
.PHONY: mockgen
mockgen:
	go run github.com/golang/mock/mockgen -source=internal/api/gmail_handler.go -destination=internal/api/mocks/mock_gmail_service.go -package=mocks GmailServiceInterface

# Tidy go.mod and go.sum (idempotent)
.PHONY: tidy
tidy:
	go mod tidy

test-db-integration:
	go test -tags=integration ./internal/data/


# DB Migration Notes:
# - For local/dev, use: make dev-up or make psql-migrate-up
# - For CI/CD, ensure migrations are present and use your pipeline's preferred method.
# - docker-migrate-up/down targets have been removed for reliability. See migrations/README.md for details.
