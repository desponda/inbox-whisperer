# Core Data Schema & Initialization

## Implementation Plan

This feature sets up the foundational data layer for Inbox Whisperer, ensuring a scalable, maintainable, and production-ready backend. It includes the SQL schema, database initialization, and supporting scripts/configuration for development.

### Goals
- Establish a modular, extensible PostgreSQL schema
- Enable easy local development and onboarding
- Lay the groundwork for clean backend architecture and future microservices

## Checklist

- [ ] Review and finalize `schema.sql` for core tables and relationships
- [ ] Add or update `docker-compose.yml` to run PostgreSQL for local development
- [ ] Add a script or Makefile target to initialize the database from `schema.sql`
- [ ] Document setup and initialization steps in `README.md`
- [ ] Scaffold Go backend directories:
    - `/internal/models` for data structs
    - `/internal/data` for DB access
    - `/internal/api` for REST handlers
    - `/cmd/server` for entrypoint
- [ ] Start an OpenAPI spec for backend endpoints
- [ ] Track progress in this feature file

## Notes
- All changes should align with the principles in `developing.md` (RESTful, clean code, separation of concerns, OpenAPI-driven)
- This feature is foundational for all future development
