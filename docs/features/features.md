# Features

## Theme & UI Principles
- All features and UI screens now use the new modern, shiny dark gradient theme (see Home page in the `web` directory for reference).
- DM Sans is the default font throughout.
- Layouts are centered, bold, visually modern, and follow SaaS best practices.
- The UI app directory has been renamed from `mvp-ui` to `web` for clarity and best practice.
- The dev workflow uses Makefile targets (`dev-up`, `dev-down`, `install`, `start`, `setup`) with the `web` directory. See Makefile for details.
- All configuration is via `config.json.template` and `.env` for environment overrides.
- The stack is orchestrated with Docker Compose (db, backend, frontend).
- All docs and configs are now aligned with this workflow.
- `.gitignore` now covers Node/React, font cache, and Python artifacts.
- A reusable UI template and setup instructions are available at `/other/README-ui-template.md` for future projects.

**Update 2025-04-24:**
- The Home page and all UI now match a premium SaaS look (Thunderbird/Linear-inspired), with a shiny dark gradient, DM Sans, and accent color.
- All documentation and configs are aligned with the new structure and workflow.

## Active Feature

### MVP React UI (2025-04-22)
See `features/mvp-ui-react.md` for implementation plan and checklist.

## Paused Features

### Gmail Integration & AI Categorization
See `features/gmail-integration-ai-categorization.md` for details.

## Completed Features

### Backend Refactoring & Testability Improvements (2025-04-22)
- Refactored Gmail message fetching and caching logic for clarity, maintainability, and testability.
- Introduced dependency injection for Gmail API, enabling easy mocking and robust testing.
- Standardized error handling for Gmail 404s with a dedicated helper (`isNotFoundError`).
- Ensured consistent extraction of the Date field from both cached and freshly fetched messages.
- All changes adhere to Go best practices and the project's design principles as described in `developing.md`.

### 1. OAuth2 Google Login & User Creation
- Users can sign up and log in using Google OAuth2.
- On first successful OAuth login, a user is created in the database and their token is stored.
- On subsequent logins with the same Google account, the user is not duplicated, and their information is not overwritten.
- Direct creation of users via the API (`POST /users/`) is forbidden; only OAuth flow can create users.
- Integration tests enforce these requirements.

### 2. Session Handling
- Session tokens are set and validated using secure cookies.
- Session simulation in tests uses the same middleware as production, ensuring accurate authentication logic.

### 3. User API Security
- All user management endpoints (`/users/`) are protected; only admin logic (future) will allow direct creation or listing.
- Non-admin attempts to create, list, or update users via API are rejected (403 Forbidden).

### 4. Integration Test Coverage
- Tests verify OAuth login, user creation, session handling, and prevention of duplicate users.
- Tests ensure that direct API user creation is forbidden.


## What to Do With Completed Features?
- Move completed features to a "Released" or "Changelog" section in your documentation (e.g., `CHANGELOG.md`), or keep them here for historical visibility.
- Optionally, tag completed features with release versions or dates.
- Archive or collapse old/completed features to keep the list focused on upcoming work.
- Use checkboxes for in-progress features, and move checked/completed ones to a separate section.

---

*This document is up-to-date as of 2025-04-22. Update as new features are completed or released.*
