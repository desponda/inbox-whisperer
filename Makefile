# Makefile for Inbox Whisperer

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

# Bring up the DB and apply all migrations (IDEMPOTENT, recommended for local/dev)
.PHONY: dev-up
dev-up:
	docker-compose up -d db
	for f in $(sort $(wildcard migrations/*_*.up.sql)); do \
	  echo "Applying $$f"; \
	  cat $$f | docker-compose exec -T db psql -U inbox -d inboxwhisperer 2>&1 | grep -v 'already exists' || true; \
	done

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

# Lint the codebase using golangci-lint (idempotent)
.PHONY: lint
lint:
	golangci-lint run ./...

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


# CI/CD/Production only: Apply/rollback migrations using golang-migrate Docker image
# WARNING: These targets will NOT work from inside a devcontainer unless the host's migrations folder is visible to Docker.
# The docker-migrate-up and docker-migrate-down targets have been removed
# because Docker volume mounting is unreliable in some environments (e.g., devcontainers, cloud IDEs).
# Use the psql-migrate-up target for local/dev and containerized setups.
