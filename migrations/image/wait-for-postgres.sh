#!/bin/sh
# wait-for-postgres.sh
# Usage: wait-for-postgres.sh host port
set -e
host="$1"
port="$2"

until pg_isready -h "$host" -p "$port"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - continuing"
