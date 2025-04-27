#!/bin/sh
# wait-for-db.sh
# Usage: wait-for-db.sh host port

set -e

host="$1"
port="$2"

until pg_isready -h "$host" -p "$port" -U inbox; do
  echo "Waiting for Postgres at $host:$port..."
  sleep 2
 done
 echo "Postgres is up at $host:$port"
