# syntax=docker/dockerfile:1

# Stage 1: Build the application
FROM golang:1.22.6-alpine AS builder

RUN time apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=1 GOOS=linux go build -o /docker-gs-ping

# Stage 2: Create the final runtime image
FROM alpine:latest

# Install only necessary runtime dependencies (if any)
RUN apk add --no-cache libc6-compat

WORKDIR /app

COPY --from=builder /docker-gs-ping /docker-gs-ping

COPY .env .env

COPY mssql_init.sql /app/mssql_init.sql

EXPOSE 8080

CMD ["/docker-gs-ping"]
