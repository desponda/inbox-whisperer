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

## Getting Started
### Backend
1. Install Go (>=1.22)
2. Copy `.env.example` to `.env` and fill in credentials
3. Run migrations: `make migrate-up`
4. Start backend: `make run`

### Frontend (React UI)
1. Scaffold the project in `ui/` (see `features/mvp-ui-react.md` for plan)
2. Use Makefile targets for dev, lint, test, and build:
   - `make ui-install` — install dependencies
   - `make ui-dev` — start dev server
   - `make ui-lint` — lint code
   - `make ui-test` — run tests
   - `make ui-build` — build for production
   - `make ui-generate-api-client` — generate TypeScript API client from OpenAPI spec
3. See [`features/mvp-ui-react.md`](features/mvp-ui-react.md) for workflow and progress

## Documentation
- Features: [`features.md`](features.md)
- Backend: [`developing.md`](developing.md)
- Frontend: [`ui/README.md`](ui/README.md)

## License
MIT
