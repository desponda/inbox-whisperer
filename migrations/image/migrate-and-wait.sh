#!/bin/sh
set -e

echo "[DEBUG] migrate-and-wait.sh loaded at $(date)"

# Wait for Postgres to be ready
/scripts/wait-for-postgres.sh db 5432

# Apply all migrations
for f in /migrations/*_*.up.sql; do
  echo "Applying $f"
  cat "$f" | PGPASSWORD="$PGPASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 || true
done
