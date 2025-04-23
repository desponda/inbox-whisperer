# MVP React UI Feature

This document tracks the implementation plan and checklist for the Inbox Whisperer MVP user interface, built with React and best practices.

## Current Status (2025-04-23)
- All backend endpoints are stable, tested, and documented (see test coverage improvement plan).
- OpenAPI spec is up-to-date; ready for API client generation.
- Makefile now includes targets for frontend (ui/) development, linting, testing, and API client generation.
- Ready to scaffold and implement the React UI.

## Overview
- Strict adherence to React and frontend best practices.
- Use of a mature UI/component framework (MUI, Chakra UI, or Ant Design).
- Auto-generated API client from OpenAPI spec.
- Clean, scalable, and maintainable code structure.
- Full TypeScript, ESLint, Prettier, and testing setup.

## Implementation Plan

### 1. Project Setup & Standards
- [ ] Scaffold a new React project (Vite or Create React App, with TypeScript).
- [ ] Configure Prettier, ESLint (Airbnb or recommended config), and TypeScript strict mode.
- [ ] Set up a UI/component framework for consistent, accessible design.
- [ ] Organize folder structure for scalability: `src/components`, `src/pages`, `src/api`, `src/hooks`, `src/types`.

### 2. API Client
- [ ] Ensure OpenAPI spec is up-to-date for all backend endpoints.
- [ ] Use openapi-generator or similar to auto-generate a TypeScript client.
- [ ] Integrate the generated client for all API calls.

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

---
*This feature is active as of 2025-04-23. Update as progress is made.*
