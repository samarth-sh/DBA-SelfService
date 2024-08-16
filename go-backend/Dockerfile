# syntax=docker/dockerfile:1

FROM golang:1.22.6-alpine

# Set destination for COPY
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download


COPY *.go ./

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -o /docker-gs-ping

EXPOSE 8080

# Run
CMD ["/docker-gs-ping"]