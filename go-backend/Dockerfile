# syntax=docker/dockerfile:1

FROM golang:1.22.6-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/go-backend/go-backend

FROM alpine:latest

RUN apk add --no-cache libc6-compat

WORKDIR /app

COPY --from=builder /app/go-backend/go-backend /app/go-backend/

COPY .env .env
COPY MSSQL-SP /app/MSSQL-SP
COPY MSSQL-UDF /app/MSSQL-UDF

EXPOSE 8080

CMD ["/app/go-backend/go-backend"]
