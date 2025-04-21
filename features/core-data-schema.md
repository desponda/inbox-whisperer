# Core Data Schema & Migration System

## Current State (as of 2025-04-21)

- **Database migrations are managed via SQL files in `migrations/`**
- **Local development uses a single idempotent command:**
  ```sh
  make dev-up
  ```
  - Brings up the Postgres DB (via Docker Compose)
  - Applies all migrations using `psql` inside the DB container
  - Safe to run multiple times (idempotent: no duplicate data or errors)
- **Creating new migrations:**
  - Run `make migrate-create` to generate new `.up.sql`/`.down.sql` files
  - Edit those files to define schema changes
- **All migrations are idempotent:**
  - Tables and indexes use `IF NOT EXISTS`
  - Seed data uses `ON CONFLICT DO NOTHING`
- **CI/CD and production:**
  - Use the `docker-migrate-up`/`down` targets if needed (ensure migrations are visible to Docker)

## Checklist (updated)
- [x] Modular, extensible PostgreSQL schema defined in migration files
- [x] `docker-compose.yml` for local Postgres
- [x] Idempotent Makefile target for DB setup (`make dev-up`)
- [x] Clear documentation in `migrations/README.md`
- [x] Scaffold Go backend directories:
    - `/internal/models` for data structs
    - `/internal/data` for DB access
    - `/internal/api` for REST handlers
    - `/cmd/server` for entrypoint
- [x] Start an OpenAPI spec for backend endpoints (see `api/openapi.yaml`)
- [x] Track progress in this feature file

---

**The core data schema, migrations, onboarding workflow, and test infrastructure (including testcontainers for repository tests) are now complete and stable.**

- Test DB setup now programmatically creates all required tables (including user_tokens) for repository tests.
- All tests pass as of 2025-04-21.

➡️ Next steps for backend API and service development are tracked in a new feature file: `features/go-backend-api.md`.

## Notes
- All changes should align with the principles in `developing.md` (RESTful, clean code, separation of concerns, OpenAPI-driven)
- This workflow ensures all contributors have a robust, reproducible, and hassle-free DB setup.
