# Web UI (React)

## Theme & Design Source of Truth
- All MVP UI must use the modern, shiny dark gradient theme, DM Sans font, and centered SaaS hero layout.
- The Home page is the reference design for all future UI screens.

## MVP React UI Feature

This document tracks the implementation plan and checklist for the Inbox Whisperer MVP user interface, built with React and best practices.

## Current Status (2025-04-24)
- All backend endpoints are stable, tested, and documented (see test coverage improvement plan).
- OpenAPI spec is up-to-date; ready for API client generation.
- Makefile now includes targets for frontend (ui/) development, linting, testing, and API client generation.
- Jest config is robust: CSS imports are properly mocked using identity-obj-proxy (must be installed as a dev dependency).
- React context/provider usage in tests is correct (UserProvider wraps components as needed).
- All frontend and backend tests pass locally (make ci, npm test).
- Next: Spin up the app locally and verify UI/UX in the browser.

## Overview
- Strict adherence to React and frontend best practices.
- Use of a mature UI/component framework (MUI, Chakra UI, or Ant Design).
- Auto-generated API client from OpenAPI spec.
- Clean, scalable, and maintainable code structure.
- Full TypeScript, ESLint, Prettier, and testing setup.

## Implementation Plan

### 1. Project Setup & Standards
- [x] Scaffold a new React project (Vite + TypeScript, strict mode enabled).
- [x] Configure Prettier, ESLint (Airbnb config), and TypeScript strict mode.
- [x] Set up DaisyUI (with Tailwind v4) for consistent, accessible, and modern design.
- [x] Organize folder structure for scalability: `src/components`, `src/pages`, `src/api`, `src/hooks`, `src/types`.
- [x] Set up Docker Compose 3-tier workflow (Postgres, backend, frontend) with Makefile `dev-up`/`dev-down` and config.json.template/.env overrides.

### 2. API Client
- [x] Ensure OpenAPI spec is up-to-date for all backend endpoints.
- [x] Use openapi-generator or similar to auto-generate a TypeScript client (present in src/api/generated).
- [x] Integrate the generated client for all API calls (in progress; some manual fetches remain).

---

**Update 2025-04-26:**
- Nginx config now supports SPA fallback and proxies /api requests to the backend, enabling full-stack login and session flows.
- OpenAPI TypeScript client is generated and present, but not yet fully integrated throughout the frontend (manual fetches still used in some places).

**Next:** Begin scaffolding the Home (landing) and Login pages using strict best practices and DaisyUI.

### 3. Authentication & Session
- [ ] Implement Google OAuth2 login flow (trigger backend endpoint, handle redirects).
- [ ] Store session/token securely (prefer httpOnly cookies; use React context for auth state).

### 4. Core UI Screens
- [ ] Login Page (Google login button, clean design).
- [ ] Inbox Page (list emails, loading/error states, pagination if needed).
- [ ] Email Detail Page (full content of selected email).
- [ ] Navigation (simple, accessible nav bar or drawer).

### 5. Best Practices & Libraries
- [ ] Use React Router for navigation.
- [ ] Use React Query (TanStack Query) or SWR for data fetching/caching.
- [ ] Use UI framework's theming and accessibility features.
- [ ] Write unit tests for components (Jest + React Testing Library).
- [ ] Use env vars for API URLs/config.

### 6. Documentation & OpenAPI
- [ ] Update OpenAPI spec as backend evolves.
- [ ] Document frontend setup and contribution in README.md.

## Checklist
- [x] Backend endpoints documented and OpenAPI up-to-date
- [ ] React project scaffolded with strict standards
- [ ] Auto-generated API client integrated
- [ ] Core UI screens implemented
- [ ] Codebase fully linted and tested
- [ ] Makefile and developer scripts updated for fullstack workflow
- [x] Backend: Add sessionCookieSecure flag to values.yaml, config.json, and Go config. Session manager now uses this flag for all cookies. Enables local dev over HTTP (set to false) and should be true in production.

## Developer Experience

- To run all checks locally (lint, typecheck, tests for backend and frontend), use `make ci` before PR or deploy.

---
*This feature is active as of 2025-04-23. Update as progress is made.*

### Frontend Session Management Refactor (2025-XX-XX)

#### Background & Best Practices
- Backend now uses secure, httpOnly cookies for session management (stateless, server-validated).
- Frontend should treat the session as opaque: no JWT parsing, no localStorage for tokens, always use `credentials: 'include'` for API calls.
- The canonical source of user/session state is `/api/users/me`.
- All protected routes/components should rely on a single React context for user/session state.
- Use SWR for caching and revalidation of user/session state (already installed).
- Logout should clear all browser storage and cookies, and redirect to login.
- All API calls should use the generated OpenAPI client.

#### Checklist: Frontend Session Management Refactor

1. **User Context & Session State**
   - [x] Refactor `UserContext` to use SWR for `/api/users/me` instead of manual fetch.
   - [x] Ensure all session state is derived from the SWR cache.
   - [x] Remove any manual fetches for user/session state elsewhere in the app.

2. **API Client Integration**
   - [x] Replace all manual fetches with the generated OpenAPI client (especially for `/api/users/me`).
   - [x] Ensure all API calls use `withCredentials`/`credentials: 'include'` as needed.

3. **Authentication Flow**
   - [x] Ensure login flow redirects to `/api/auth/login` and handles callback via `/auth/callback`.
   - [x] On successful callback, revalidate user context/session state (ensure SWR mutate is called after callback if not already present).
   - [x] On 401 from any API call, trigger session expiry logic (clear state, redirect to login).

4. **Protected Routes**
   - [x] Ensure all protected pages/components use `ProtectedRoute` and rely on `UserContext` for auth state.
   - [x] Show a loading spinner while session state is being determined.

5. **Logout**
   - [x] Implement a logout button/component that:
     - [x] Clears all local/session storage and cookies.
     - [x] Resets user context and redirects to login.
     - [ ] Calls a backend logout endpoint (if available; not required for MVP).

6. **Error Handling**
   - [x] Show user-friendly messages for session expiry, auth errors, and forbidden access.
   - [x] Ensure all error states are handled gracefully in the UI.

7. **Testing**
   - [x] Update or add tests for:
     - [x] Session expiry and re-login flow.
     - [x] Protected route access.
     - [x] User context revalidation after login/logout.

8. **Documentation**
   - [x] Document the new session management flow in this file.
   - [x] Add developer notes on how to use the new context/hooks and how to test session flows.

#### Session Management Flow Summary
- The canonical source of user/session state is `/api/users/me`, fetched via SWR in `UserContext`.
- All protected routes use `ProtectedRoute` and the `useUser` hook for session state.
- Login and logout flows update the session state via SWR's `mutate`.
- Session expiry is handled globally: on 401, the user is logged out and redirected to login.
- All API calls use the generated OpenAPI client with `withCredentials` enabled.
- To test session flows, use the tests in `UserContext.test.tsx` and simulate API responses as needed.

#### Library Recommendation
- **SWR** (already installed) is the recommended library for session/user state: simple, lightweight, and React-friendly.
- No need for Redux or heavyweight state management—React context + SWR is best practice for session/user state in modern React apps.

#### Sample SWR Integration for User Context
```tsx
import useSWR from 'swr';
import { getUser } from '../api/generated/user/user';

export function useUser() {
  const { data, error, isLoading, mutate } = useSWR('/api/users/me', () =>
    getUser().getApiUsersMe({ withCredentials: true }).then(res => res.data)
  );
  // ...return user, loading, error, logout, etc.
}
```

- [x] Backend: Add sessionCookieSecure flag to values.yaml, config.json, and Go config. Session manager now uses this flag for all cookies. Enables local dev over HTTP (set to false) and should be true in production.
- [x] Remove any manual fetches for user/session state elsewhere in the app.
- [ ] Ensure backend does **not** set `Secure` flag on cookies in dev (HTTP). If you want to match production, add local HTTPS ingress with self-signed cert for kind and set `Secure` flag always.
