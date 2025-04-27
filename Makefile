# Makefile for Inbox Whisperer

.PHONY: install start setup help clean lint vet staticcheck lint-strict test ci tidy ui-install ui-dev ui-build ui-lint ui-test ui-typecheck ui-coverage ui-generate-api-client kind-up kind-load dev-deploy dev-up dev-down backend-build frontend-build format security

# Install all dependencies and fonts
install:
	npm install --prefix web

# Build all images (backend, frontend, migrate)
build-all:
	bash scripts/dev/dev-deploy.sh --build-all

# Build backend image and restart deployment
build-backend:
	bash scripts/dev/dev-deploy.sh --build-backend

# Build frontend image and restart deployment
build-frontend:
	bash scripts/dev/dev-deploy.sh --build-frontend

# Build migration image
build-migrate:
	bash scripts/dev/dev-deploy.sh --build-migrate

# Start the dev server
start:
	npm run dev --prefix web -- --host

# Format Go and frontend code
format:
	gofmt -w .
	npm run fmt --prefix web || true

# Security checks (example: gosec)
security:
	gosec ./...

# Full setup (install + start)
setup: install start
#
# Usage:
#   make <target>
#
# Key targets:
#   ci                  Run all backend and frontend lint, typecheck, tests, and builds (canonical local CI)
#   backend-build       Build backend Docker image (uses cmd/backend/.dockerignore)
#   frontend-build      Build frontend Docker image (uses web/.dockerignore)
#   format              Format all code (Go + frontend)
#   security            Run security checks (Go)
#   dev-up              Start full stack for local dev (docker-compose, kind, or scripts)
#   dev-down            Tear down local dev stack
#   ...                 (see below for more targets)
#
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_-]+:|^# ' Makefile | \
	  sed -E 's/^([a-zA-Z0-9_-]+):.*/\1/;s/^# //' | \
	  awk 'BEGIN{t=""} /^[a-zA-Z0-9_-]+$/ {if(t!=""){print t}; t=$$0; next} {t=t" "$0} END{print t}' | \
	  column -t -s ':'

clean:
	rm -f test-output.txt

# Go targets
lint:
	@echo 'Checking gofmt...'
	@gofmt -l . | grep -v '^vendor/' | tee /tmp/gofmt.out
	@if [ -s /tmp/gofmt.out ]; then echo 'gofmt needs to be run on these files:'; cat /tmp/gofmt.out; exit 1; fi
	golangci-lint run ./...

# Frontend targets
ui-install:
	npm install --prefix web

ui-dev:
	npm run dev --prefix web -- --host

ui-build:
	npm run build --prefix web

ui-lint:
	npm run lint --prefix web

ui-test:
	npm test --prefix web

ui-typecheck:
	npm run typecheck --prefix web

ui-coverage:
	npm run coverage --prefix web

ui-generate-api-client:
	npm run generate:openapi --prefix web

# KIND/Kubernetes targets
kind-up:
	kind create cluster --name inbox-whisperer || true

kind-load:
	kind load docker-image inboxwhisperer-backend:latest --name inbox-whisperer
	kind load docker-image inboxwhisperer-frontend:latest --name inbox-whisperer

# Deploy all (build, load, secrets, helm upgrade/install, restart deployments)
deploy-all:
	bash scripts/dev/dev-deploy.sh --deploy-all

# Local dev up/down (example: using scripts or docker-compose)
dev-up:
	bash scripts/dev/dev-deploy.sh

create-secrets:
	bash scripts/dev/dev-deploy.sh --create-secrets

# DB migration
migrate-create:
	@read -p "Migration name: " name; migrate create -dir ./migrations -ext sql $$name

.PHONY: install start setup help clean lint vet staticcheck lint-strict test ci tidy ui-install ui-dev ui-build ui-lint ui-test ui-typecheck ui-coverage ui-generate-api-client kind-up kind-load dev-deploy dev-up dev-down backend-build frontend-build format security

ui-fmt:
	npm run fmt --prefix web

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
