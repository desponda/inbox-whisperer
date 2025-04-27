# Inbox Whisperer

Inbox Whisperer is an AI-powered tool to help users achieve and maintain inbox zero. It connects to Gmail, fetches emails, and (in future versions) categorizes them using AI. The project includes a Go backend and a modern React frontend.

## Features
- Google OAuth2 login & secure session management
- Fetch emails from Gmail via API
- Store and cache emails in a PostgreSQL database
- (Planned) AI-powered email categorization
- REST API for frontend/backend communication
- React SPA frontend with TypeScript and a UI framework

## Project Structure
```
├── api/           # OpenAPI specs and API-related code
├── cmd/           # Go entrypoints (backend server, migrations)
├── internal/      # Go backend source
├── migrations/    # SQL migrations
├── ui/            # React frontend (Vite, TypeScript)
├── features/      # Markdown specs and checklists
├── scripts/       # Dev scripts
├── developing.md  # Design, workflow, and coding principles
├── features.md    # Feature summary and status
├── CHANGELOG.md   # Release notes
```

## Development

### Local CI (Run All Checks)

To run all backend and frontend lint, typecheck, and tests locally (exactly as in GitHub Actions CI), use:

```sh
make ci
```

This is the canonical way to check your code before pushing or opening a PR.

**Frontend Jest CSS Import Troubleshooting:**
- If you see errors about CSS imports in Jest, make sure identity-obj-proxy is installed as a dev dependency and your jest.config.cjs has the correct moduleNameMapper/moduleFileExtensions settings.

## Quickstart: Full-Stack Dev Environment

1. Copy `config.json.template` to `config.json` and fill in your real credentials (never commit secrets).
2. Optionally, copy `.env.example` to `.env` and override any environment variables (API URLs, etc).
3. **For local Kubernetes (kind) development, use:**

   ```sh
   make dev-deploy
   ```
   This is the canonical, idempotent way to build, load, and deploy all components to your kind cluster. It will auto-detect or create a kind cluster, build all images, load them, and deploy via Helm using the latest code.

4. For Docker Compose-based development (without Kubernetes), use:

   ```sh
   make dev-up
   ```
   This brings up Postgres, backend, frontend, and applies all DB migrations (idempotent, canonical workflow).

5. To bring everything down and clean up volumes:

   ```sh
   make dev-down
   ```

- The main React frontend lives in `web/` (not `ui/`).
- Backend and frontend Dockerfiles are in `cmd/backend/` and `web/` respectively.
- The only scripts you should run directly are in `scripts/dev-deploy.sh` (for kind) and Makefile targets. **Other scripts in `scripts/` are required for container startup (e.g., `wait-for-db.sh`) or database migrations, and should not be deleted.**
- See the Makefile for all targets and developer scripts.
- All config is environment-variable driven for dev/staging/prod flexibility.

## Continuous Integration (CI)

All code and documentation changes are automatically checked in GitHub Actions:
- **Backend:** Linting, vet, staticcheck, tidy, build, and test
- **Frontend:** Install, lint, format, test, build
- **Docs:** All markdown files in `/docs/` are linted for style and correctness using markdownlint
- **Pull Requests:** PR titles are checked for semantic correctness

You can view and troubleshoot CI runs in the GitHub Actions tab.

## License
MIT
