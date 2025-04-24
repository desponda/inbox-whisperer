# Backend Dockerfile
FROM golang:1.21-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o inboxwhisperer ./cmd/backend

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/inboxwhisperer ./inboxwhisperer
COPY config.json.template ./config.json
COPY migrations ./migrations
EXPOSE 8080
CMD ["./inboxwhisperer"]
