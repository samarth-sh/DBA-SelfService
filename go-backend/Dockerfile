# syntax=docker/dockerfile:1

FROM golang:1.22.6-alpine

# Install necessary dependencies
RUN apk add --no-cache gcc musl-dev

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux go build -o /docker-gs-ping

EXPOSE 8080

# Run the Go application
CMD ["/docker-gs-ping"]
