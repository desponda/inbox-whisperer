# Database Migrations (Inbox Whisperer)

This folder contains all versioned database migrations for Inbox Whisperer, managed by [golang-migrate](https://github.com/golang-migrate/migrate).

## Local Development: One-Command Setup (Recommended)

**To set up or update your database for development, run:**

```sh
make dev-up
```
- Brings up the database (if not already running)
- Applies all migrations using `psql` inside the db container
- **Idempotent:** Safe to run multiple times—no errors, no duplicate data
- Works in devcontainers, VSCode, or any local setup

## Creating a Migration

```sh
make migrate-create
```
- Prompts for a migration name and creates new `.up.sql` and `.down.sql` files in the `migrations/` folder
- Edit the generated files to define your schema changes

## CI/CD and Production

- Use the `docker-migrate-up` and `docker-migrate-down` targets only in CI/CD or production environments where the migration files are present on the host filesystem.
- These targets may **not** work from inside a devcontainer.

## Best Practices

- **Never edit old migration files after they’ve been applied**—always add a new migration for changes.
- Keep migration files small and focused on a single change.
- All migrations should be idempotent (use `IF NOT EXISTS`, `ON CONFLICT DO NOTHING`, etc.) for a smooth dev experience.
- Review the [golang-migrate docs](https://github.com/golang-migrate/migrate) for advanced usage.

## Example Commands

```sh
make dev-up              # Bring up DB and apply all migrations (local/dev)
make migrate-create      # Create a new migration
make docker-migrate-up   # (CI/CD only) Apply all migrations using Docker
make docker-migrate-down # (CI/CD only) Roll back the last migration
```

---

**All schema changes must be tracked via migrations. Do not edit the database or schema manually.**
