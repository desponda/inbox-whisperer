FROM golang:1.24.2 as build

RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*

COPY scripts/prod/wait-for-postgres.sh /scripts/wait-for-postgres.sh
COPY scripts/prod/migrate-and-wait.sh /scripts/migrate-and-wait.sh
COPY migrations /workspaces/inbox-whisperer/migrations

RUN chmod +x /scripts/wait-for-postgres.sh /scripts/migrate-and-wait.sh

ENTRYPOINT ["/bin/sh", "/scripts/migrate-and-wait.sh"]
