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
- [ ] Store and refresh access tokens securely (per user).
- [ ] Document setup for local and production (redirect URIs, credentials).

### 2. Gmail API Fetching
- [ ] Implement service to fetch emails from Gmail API (using user's token).
- [ ] Store fetched emails in database (raw and/or normalized form).
- [ ] Handle pagination, rate limits, and partial syncs.

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
- [ ] Gmail OAuth2 flow implemented and tested
- [ ] Emails fetched and stored
- [ ] Categorization service working and tested
- [ ] REST endpoints exposed and documented
- [ ] All new code covered by unit/integration tests
- [ ] Docs and OpenAPI updated
