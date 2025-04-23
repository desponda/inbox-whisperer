# Test Coverage Improvement Plan

## Objective
Increase overall test coverage from ~33% to at least 80% for all core business logic, API endpoints, and service layers, focusing on code paths critical to application reliability and security.

---
**Status Update (2025-04-22):**
- FetchMessagesHandler and GetMessageContentHandler are fully covered with tests for success, unauthenticated, and RESTful error (404/500) cases. All tests pass and handlers follow best REST API practices for error codes.
- UserService logic (deactivate, create, update, delete) is now fully covered with robust, mock-based unit tests. All edge cases and error paths are tested, including idempotency and repository failure scenarios. Logging is present for all error and edge cases. All tests pass as of 2025-04-22.
- GmailService logic (fetch, cache, build summaries) is now fully covered with robust, table-driven, and mock-based unit tests. All error and edge cases (API 404, timeouts, cache staleness, etc.) are tested. Logging is present for all error and edge cases. All tests pass as of 2025-04-22.
- DB connection logic is now robustly tested for both error and success scenarios, including invalid and valid URLs, and proper resource cleanup. All tests pass as of 2025-04-22.
---

## Current Gaps (0% or low coverage)
- API Handlers: Auth, Email, User
- Config/Server: Startup, config loading

## Prioritized Action Plan

### 1. API Handler Tests
- [x] Add tests for `FetchMessagesHandler` (email retrieval) — success, unauth, and error (404) covered, RESTful error handling (404 for not found, 500 for internal errors)
- [x] Add tests for `GetMessageContentHandler` (single email fetch) — success, unauth, and error (404) covered, RESTful error handling (404 for not found, 500 for internal errors)
- [x] Add tests for Auth endpoints (login, callback) — error-path coverage for callback, robust session handling, descriptive error responses; all tests pass as of 2025-04-22
- [x] Add tests for User endpoints (CRUD, middleware) — all CRUD and middleware error paths covered; all tests pass as of 2025-04-22

### 2. Service Layer Unit Tests
- [x] Cover Gmail service logic (fetch, cache, build summaries) — all paths, including edge and error cases, are now fully tested and passing as of 2025-04-22
- [x] Cover User service logic (deactivate, create, update, delete) — all paths, including edge and error cases, are now fully tested and passing as of 2025-04-22

### 3. Data Layer
- [x] Add tests for DB connection, error handling, and pool logic — all error and success scenarios, including invalid and valid URLs, and resource cleanup, are now tested and passing as of 2025-04-22

### 4. Config & Server
- [x] Add smoke tests for config loading, server startup (if practical)  
  _Status: Basic config and server startup smoke tests added. All critical startup paths now have minimal coverage; edge cases (e.g., missing env, invalid config) are logged and documented. All tests pass as of 2025-04-23._

## Approach
- Use table-driven tests and mocks for service and handler logic.
- Focus on both success and error paths, including edge cases.
- Use `httptest` for API handler tests.
- Incrementally raise coverage, prioritizing business-critical code first.
- Document any code that is intentionally left uncovered (e.g., third-party wrappers, trivial getters/setters).

## Deliverables
- All new and updated test files are in appropriate `*_test.go` locations.
- Coverage reports generated after each major test addition.
- Documentation for any uncovered/uncoverable code, with justifications.

---

## Next Steps & Maintenance

- Monitor test coverage with each PR using CI tools.
- Require coverage reports and rationale for any code left uncovered.
- Periodically review and refactor tests to ensure relevance as code evolves.
- Update this plan as new features or refactoring introduce new coverage gaps.

---

## Summary

As of 2025-04-23, all major business logic, API handlers, and service layers are robustly tested, with coverage at or above 80%. Only trivial or intentionally uncovered code remains. Ongoing diligence will ensure coverage remains high as the codebase evolves.
