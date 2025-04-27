FROM alpine:3.18

# Install Postgres client tools
RUN apk add --no-cache postgresql-client

# Copy migrations and scripts
COPY ./*.sql /migrations/
COPY migrate-and-wait.sh wait-for-postgres.sh /scripts/

# Ensure scripts are executable
RUN chmod +x /scripts/*.sh

# Entrypoint
ENTRYPOINT ["/scripts/migrate-and-wait.sh"]
