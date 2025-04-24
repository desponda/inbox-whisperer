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
3. Start everything (Postgres, backend, frontend, migrations) with:

   ```sh
   make dev-up
   ```
   This is idempotent and applies all DB migrations.

4. To bring everything down and clean up volumes:

   ```sh
   make dev-down
   ```

- See the Makefile for all targets and developer scripts.
- The main React frontend lives in `web/` (not `ui/`).
- Backend and frontend Dockerfiles are in `cmd/backend/` and `web/` respectively.
- All config is environment-variable driven for dev/staging/prod flexibility.

## Documentation
- Features: [`features.md`](features.md)
- Backend: [`developing.md`](developing.md)
- Frontend: [`ui/README.md`](ui/README.md)

## License
MIT
