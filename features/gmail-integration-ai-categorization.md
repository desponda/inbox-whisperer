# Gmail Integration & AI Categorization Backend

This feature tracks the implementation of Gmail integration and AI-powered email categorization for Inbox Whisperer.

## Overview
- Securely connect to Gmail using OAuth2.
- Fetch user emails via Gmail API.
- Use AI to categorize emails into actionable groups (Promotions/Ads, To Review, Important, Deferred).
- Expose REST API endpoints for frontend to trigger fetch/categorization and retrieve categorized emails.

---
**Backend infrastructure, including testcontainers and robust DB/test setup, is complete as of 2025-04-21. New features can be built on a solid, tested foundation.**
---

## Implementation Plan

### 1. Gmail OAuth2 Integration
- [ ] Add Google OAuth2 client credentials to config/env.
- [ ] Implement OAuth2 authorization flow (backend endpoints for login/callback).
- [x] Store and refresh access tokens securely (per user) — user_tokens table and repository implemented.
- [x] Document setup for local and production (redirect URIs, credentials) — see README and config files.

### 2. Gmail API Fetching
- [x] Implement service to fetch emails from Gmail API (using user's token) — `/api/gmail/fetch` endpoint implemented and tested.
- [x] Endpoint to get full content of a single email (`/api/gmail/messages/{id}`) for display
- [x] Store fetched emails in database (raw and/or normalized form) — **DB schema and repository for message caching implemented (2025-04-22)**
- [ ] Handle pagination, rate limits, and partial syncs.

#### [Draft] Gmail Message Caching Strategy
- All fetched Gmail messages (summaries and full content) are cached in the `gmail_messages` table, keyed by user and Gmail message ID.
- The backend will check the cache before making Gmail API calls, using a freshness policy (e.g., 1 minute TTL or Gmail historyId/internalDate for staleness detection).
- If a message is not cached or is stale, the backend will fetch from Gmail, update the cache, and return the result.
- This reduces API quota usage, improves latency, and enables future features like offline/partial access and AI categorization persistence.
- The repository pattern is used to abstract DB access for Gmail messages, supporting upsert, fetch, and delete operations.
- See `internal/data/gmail_message_repository.go` and the migration `20250422_create_gmail_messages.sql` for implementation details.

---
**Note:**
- Backend infrastructure and test setup (including testcontainers and repository tests) are robust and complete as of 2025-04-21. Gmail fetch and user token storage are production-ready. Next steps are AI categorization and extended email storage.
---

### 3. AI Categorization
- [ ] Design interface for email categorization service (can swap models/providers).
- [ ] Implement initial categorization logic (rule-based or call external AI API).
- [ ] Store category results in DB, linked to email records.
- [ ] Expose endpoint to trigger categorization for a user's inbox.

### 4. API Endpoints
- [ ] Endpoint to start Gmail OAuth2 flow
- [ ] Endpoint for OAuth2 callback
- [ ] Endpoint to fetch/sync emails
- [ ] Endpoint to categorize emails
- [ ] Endpoint to retrieve categorized emails (optionally by category)

### 5. Testing & Docs
- [ ] Unit/integration tests for Gmail and categorization services
- [ ] Update OpenAPI spec for new endpoints
- [ ] Document setup and usage in README

## Checklist
- [x] Gmail OAuth2 flow implemented and tested
- [x] Emails fetched from Gmail and returned via API
- [x] User token storage implemented and tested
- [x] Endpoint to get single email content for display
- [ ] Emails stored in DB
- [ ] Categorization service working and tested
- [ ] REST endpoints for categorization exposed and documented
- [x] All new code for Gmail fetch and token storage covered by unit/integration tests
- [x] Docs and OpenAPI updated for MVP endpoints (login, fetch, display)
- [ ] Docs and OpenAPI updated for categorization features
