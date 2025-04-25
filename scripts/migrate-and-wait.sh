#!/bin/sh
set -e

# Wait for Postgres to be ready
/scripts/wait-for-postgres.sh db 5432

# Apply all migrations
for f in /workspaces/inbox-whisperer/migrations/*_*.up.sql; do
  echo "Applying $f"
  cat "$f" | PGPASSWORD=inboxpw psql -h db -U inbox -d inboxwhisperer -v ON_ERROR_STOP=1 || true
done
