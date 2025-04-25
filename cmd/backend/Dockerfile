# Backend Dockerfile
# Use latest Go as of April 2025
FROM golang:1.24.2 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# For development: optionally bake config.json into the image
ARG DEV_CONFIG=false
COPY config.json /app/config.json
RUN CGO_ENABLED=0 go build -o inboxwhisperer ./cmd/server

# Use latest Alpine as of April 2025
FROM alpine:3.19.1
WORKDIR /app
COPY --from=build /app/inboxwhisperer ./inboxwhisperer
COPY migrations ./migrations
EXPOSE 8080
# Add labels for best practices
LABEL org.opencontainers.image.source="https://github.com/desponda/inbox-whisperer"
LABEL org.opencontainers.image.description="Inbox Whisperer backend service"

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s CMD wget -q --spider http://localhost:8080/health || exit 1

CMD ["./inboxwhisperer"]
