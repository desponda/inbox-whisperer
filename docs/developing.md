# Developing.md

## Project Overview

Inbox Whisperer is a suite of AI-powered tools designed to help users achieve and maintain inbox zero. Our first feature is the Inbox Zero Helper—a guided workflow that enables users to efficiently triage and organize their Gmail inbox, ensuring that only truly important emails remain.

## Design System & Theme
- **Theme:** All UI pages must use a modern, shiny dark gradient background (radial/linear blend of #0e1015, #181a20, #23243a, with blue/cyan highlight for shine).
- **Font:** Use DM Sans throughout the app. Import via `web/public/fonts/dmsans.css` and set as the default font-family.
- **Layout:** Hero and main sections should be centered, with bold, modern headlines and accent color highlights. Use the Home page as the reference for all future pages.
- **Accent Color:** Use the accent color (`#14e0c9`) for highlights, buttons, and gradients.
- **No glassmorphism or excessive shadows.**

## Running the Full Stack Locally

- Copy `config.json.template` to `config.json` and fill in real values (never commit secrets).
- Create a `dev.env` file (see `.env.example` or the template provided) to override API URLs, ports, database credentials, and other secrets for local development. This file will be loaded automatically by `dev-deploy.sh` and Makefile workflows.
- Use the Makefile:
  - `make dev-deploy` is the canonical, idempotent workflow for local Kubernetes (kind) development. It will build, load, and deploy all components to your kind cluster, auto-detecting or creating the cluster as needed. This is the only script you should ever need for local kind development.
  - `make dev-up` brings up Postgres, backend, frontend, and applies all DB migrations (idempotent, canonical workflow for Docker Compose-based development).
  - `make dev-down` brings down all containers and cleans up volumes.
- All configuration is environment-variable driven (see Docker Compose and Makefile for details).

- The only scripts you should run directly are in `scripts/dev-deploy.sh` (for kind) and Makefile targets. **Other scripts in `scripts/` are required for container startup (e.g., `wait-for-db.sh`) or database migrations, and should not be deleted.**
- All additional documentation (feature specs, development guides, migration notes, etc.) is now located in the `/docs/` directory for clarity.
- Main project README is at the project root as always.

## MVP Focus (2025-04-22)
- Users sign up/log in via Google OAuth2
- After login, users can fetch their emails (list)
- Users can view the content of a specific email
- All endpoints are described in OpenAPI and the spec is always kept up-to-date
- **Frontend is being built as a React SPA, using strict best practices, a UI framework, and an auto-generated OpenAPI client.**

## Core Workflow

1. **Email Fetching**: The app securely connects to Gmail and fetches the user's emails.

### Email Fetching Best Practices
- **Inbox loads instantly from DB cache:** When a user loads the inbox, the backend returns cached email summaries from the database for a fast, snappy experience.
- **Background sync:** Simultaneously, the backend triggers a background sync with Gmail to fetch new/updated summaries (using the Gmail API's list/metadata endpoints). The cache is updated in the background.
- **Summaries vs. Full Content:** The summary endpoint (`FetchMessages`) returns only minimal fields (subject, sender, snippet, date, etc.)—never the full body/content. The full content is fetched only via a separate endpoint (`FetchMessageContent`) and only for the selected message.
- **Client updates:** After sync, the client can poll or receive a push notification to refresh the inbox view with new data.

2. **Categorization**: Each email is automatically categorized using AI into one of four actionable groups:
   - **Promotions/Ads (FYI)**: Marketing emails, advertisements, and newsletters—content the user likely doesn't want to open or act on.
   - **To Review**: Emails the user probably wants to read, but that do not require a response (e.g., updates, reports, notifications).
   - **Important**: Emails that are personal, time-sensitive, or require a response or action.
   - **Deferred**: Emails that require action or follow-up at a later time.

   (Note: These categories are designed to be intuitive and actionable. If you have suggestions for more effective categories, please document them here.)

3. **Guided Cleanup**:
   - Users first review "Promotions/Ads (FYI)" emails, with the option to bulk archive.
   - Next, users process "To Review" emails, choosing to archive, defer, or mark as "Important".
   - Then, users review "Deferred" emails to schedule follow-up actions.
   - Finally, users review "Important" emails to confirm their status or take necessary action.

4. **Desired Outcome**: The goal is for users to either archive or appropriately categorize every email, leaving only genuinely important or deferred items in their inbox.

## Design and Coding Principles

### Backend
- **RESTful API**: The backend will be designed as a strongly RESTful API using the go-chi framework for routing and middleware.
- **Separation of Concerns**: Code will be organized to ensure clear separation between business logic, data access, and presentation layers.
- **Clean Code**: We will follow clean coding principles—readable, maintainable, and well-documented code.
- **Interfaces & Abstraction**: Use interfaces and abstract patterns where appropriate to enable flexibility, testing, and future extension.

### Frontend
- **React + Vite + TypeScript**: The frontend is built as a strict, maintainable React SPA using Vite and TypeScript (strict mode enabled).
- **UI Framework**: DaisyUI (on Tailwind v4) is used for all components to ensure accessible, modern, and consistent design. No custom CSS unless absolutely necessary—leverage DaisyUI and Tailwind utilities.
- **Folder Structure**: All code is organized for scalability and maintainability:
  - `src/components/` – Reusable UI components
  - `src/pages/` – Top-level route/page components (e.g., Home, Login, Inbox)
  - `src/api/` – Auto-generated API client and related code
  - `src/hooks/` – Custom React hooks
  - `src/types/` – TypeScript types and interfaces
- **Linting & Formatting**: Prettier and ESLint (Airbnb config, React, TypeScript, Prettier integration) are strictly enforced. No sloppy or inconsistent code is tolerated.
- **Routing**: Uses React Router for all navigation. No ad-hoc routing logic.
- **OpenAPI Integration**: The OpenAPI spec is always kept up-to-date and used to auto-generate the TypeScript API client. No manual API types or fetch logic.
- **Best Practices Only**: All code must be clean, modular, and easy to edit or extend in the future. No shortcuts or legacy patterns. All changes are tracked in `features/mvp-ui-react.md` for transparency and ongoing improvement.
- **Living Checklist**: The implementation plan and progress for the UI is always reflected in `features/mvp-ui-react.md`. Any deviation from best practices must be justified and documented.

### API Contract
- **OpenAPI Spec**: All API endpoints will be defined with OpenAPI. This ensures a strong, always-updated contract between frontend and backend, and enables auto-generation of the JavaScript client for React.
- **Auto-Generated Clients**: The OpenAPI spec will be used to generate and update the frontend API client, reducing manual work and preventing contract drift.

## Feature Development Workflow

- Each feature will have its own Markdown file within a dedicated `features/` directory.
- Each feature file will include:
  - An implementation plan.
  - A checklist of tasks for AI tools to read, update, and track progress.
- Workflow for new features:
  1. The AI will first create an implementation plan for the feature.
  2. The AI will generate a checklist of actionable tasks.
  3. The AI will proceed with development, updating the checklist as progress is made.
  4. After each development session, changes will be committed to checkpoint progress and maintain context for future sessions.
- This approach ensures persistent, up-to-date context for both humans and AI tools, supporting efficient, incremental development.

## Development Notes

- The Makefile has been improved: new `clean` and `help` targets, docker-migrate-up/down removed, and DB migration workflow clarified (see Makefile and README).
- Large/test output files are now ignored and removed from git (see .gitignore).
- Use `make help` to discover all Makefile targets.
- This document is maintained by the Windsurf context tracker.
- Please update with any new insights, workflow changes, or category refinements as development progresses.
