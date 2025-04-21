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

test-db-integration:
	go test -tags=integration ./internal/data/

# CI/CD/Production only: Apply/rollback migrations using golang-migrate Docker image
# WARNING: These targets will NOT work from inside a devcontainer unless the host's migrations folder is visible to Docker.
.PHONY: docker-migrate-up
docker-migrate-up:
	docker-compose up -d db
	docker run --rm -v "$(PWD)/migrations:/migrations" --network inbox-whisperer_default migrate/migrate \
	  -path=/migrations -database "postgres://inbox:inboxpw@db:5432/inboxwhisperer?sslmode=disable" up

.PHONY: docker-migrate-down
docker-migrate-down:
	docker-compose up -d db
	docker run --rm -v "$(PWD)/migrations:/migrations" --network inbox-whisperer_default migrate/migrate \
	  -path=/migrations -database "postgres://inbox:inboxpw@db:5432/inboxwhisperer?sslmode=disable" down 1
