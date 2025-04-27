#!/bin/bash
# Initialize the database using schema.sql
set -e

docker-compose up -d db
# Wait for Postgres to be ready
until pg_isready -h localhost -U inbox; do
  echo "Waiting for database..."
  sleep 1
done
# Copy schema.sql into the container and run it
cat schema.sql | psql -h localhost -U inbox -d inboxwhisperer
