# Go Backend API & Service Layer

This feature tracks the next phase of backend development for Inbox Whisperer, following the principles in `developing.md`:
- Always working, tested code
- Separation of concerns (API, service, data)
- Structured logging and observability
- Modern, maintainable Go idioms

## Goals
- Use [go-chi](https://github.com/go-chi/chi) for routing/middleware
- Replace gorilla/mux in main.go
- Add logging middleware and structured logging everywhere
- Scaffold config loading from environment
- Set up DB connection (pgx)
- Service layer for business logic
- Thin, validated API handlers
- Full test coverage for services and handlers

## Step-by-Step Plan

### 1. Replace gorilla/mux with go-chi
- [x] Remove gorilla/mux dependency
- [x] Add go-chi to go.mod
- [x] Refactor `cmd/server/main.go` to use chi.Router
- [x] Implement `/healthz` endpoint
- [x] Add logging middleware using zerolog
- [x] Add basic integration test for healthz

### 2. Configuration & Environment
- [x] Create `pkg/config` for env/config loading
- [x] Add `.env.example` and update README
- [x] Test config loading

### 3. Database Layer
- [x] Implement DB connection using pgx (see `internal/data/db.go`)
- [x] Add repository interfaces and basic CRUD (User)
- [x] Add tests for data layer (unit + integration with testcontainers)
- [x] Integration test infra (testcontainers-go) implemented
- [x] Devcontainer updated for testcontainers-go

### 4. Service Layer
- [x] Scaffold service and repository interfaces for a sample entity (User)
- [x] Implement service structs/methods for core business logic
- [ ] Integrate service layer into server
- [ ] Add unit tests for services

---
**Next step:**
- Rebuild devcontainer and verify integration tests work (see README for troubleshooting testcontainers in devcontainers)

---

**Note:**
- The data layer is currently Postgres-specific, but the service layer will depend on interfaces for future flexibility and testability. This keeps the app easy to refactor or extend to other backends later.
### 5. API Layer
- [ ] Implement HTTP handlers (thin, just call service)
- [ ] Add request validation
- [ ] Add tests for handlers

### 6. CI & Quality
- [ ] Add Makefile targets for lint, test, and build
- [ ] Set up GitHub Actions for CI

---

**All code must build and pass tests at every step.**

Progress is tracked here as we build out the production-ready Go backend.
