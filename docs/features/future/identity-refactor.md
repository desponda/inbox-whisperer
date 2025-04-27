# Identity Refactor Plan

## Context
Currently, the app uses the Google user ID as the primary key in the `users` table (id as TEXT) for rapid development. This is a temporary solution for development only.

## Problem
- The current approach ties internal user identity to Google (or other provider) IDs, making it hard to support multiple providers, account linking, or robust user management.
- All internal references are to an external ID format, creating tech debt.

## Solution (Best Practice)
- Refactor to use an internal UUID as the primary key for `users`.
- Introduce a `user_identities` table to map external provider accounts (Google, GitHub, etc.) to internal users.
- All app logic should use the internal UUID for user references.
- On login, look up or create a user via the `user_identities` table, not by external ID.

### Example Schema
```
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_user_id)
);
```

## Migration Plan
1. Add the new `users` and `user_identities` tables.
2. Backfill existing users into the new structure.
3. Update all references and logic to use internal UUIDs.
4. Remove the temporary shortcut.

## Status
- [ ] Not started (as of 2025-04-27)
- [ ] In progress
- [ ] Complete

---
**This file documents tech debt and the future plan for robust, provider-agnostic identity management.**
