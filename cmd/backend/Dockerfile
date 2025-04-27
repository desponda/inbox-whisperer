# Backend Dockerfile
# Use latest Go as of April 2025
FROM golang:1.24.2 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o inboxwhisperer ./cmd/server
RUN mkdir -p /app/config
# IMPORTANT: config.json is no longer baked into the image.
# Please provide /app/config/config.json at runtime (e.g., via docker-compose volume, Kubernetes ConfigMap/Secret, or manual copy).

# Use latest Alpine as of April 2025
FROM alpine:3.19.1
WORKDIR /app
RUN apk add --no-cache postgresql-client
COPY --from=build /app/inboxwhisperer ./inboxwhisperer
# Do NOT copy config.json from build context; it must be mounted/copied at runtime.
COPY migrations ./migrations
COPY scripts/prod/wait-for-db.sh ./wait-for-db.sh
RUN chmod +x ./wait-for-db.sh
EXPOSE 8080
# Add labels for best practices
LABEL org.opencontainers.image.source="https://github.com/desponda/inbox-whisperer"
LABEL org.opencontainers.image.description="Inbox Whisperer backend service"

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s CMD wget -q --spider http://localhost:8080/health || exit 1

CMD ["/bin/sh", "-c", "echo 'DEBUG: Listing /app/config:' && ls -l /app/config && echo 'DEBUG: Showing /app/config/config.json:' && cat /app/config/config.json && ./wait-for-db.sh db 5432 && ./inboxwhisperer"]
