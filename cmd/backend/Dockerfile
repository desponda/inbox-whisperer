# Backend Dockerfile
FROM golang:1.24.2 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o inboxwhisperer ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/inboxwhisperer ./inboxwhisperer
COPY migrations ./migrations
EXPOSE 8080
CMD ["./inboxwhisperer"]
