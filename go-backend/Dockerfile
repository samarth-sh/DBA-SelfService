
# Set the Go version and base image
ARG GO_IMAGE_VER=1.23-alpine
FROM golang:${GO_IMAGE_VER} AS builder

# Set working directory inside the container
WORKDIR /go/src/app

# Copy necessary files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Set build arguments for versioning and metadata
ARG BUILD_STAMP=$(date +%Y%m%d%H%M%S)
ARG GIT_COMMIT_ID=2f1aeb4ea80be0f9e7a627e26302244a6d9fa98a
ARG GIT_PRIMARY_BRANCH=main
ARG GIT_URL=https://github.com/angel-one/dba-self-service-go.git
ARG HOST_NAME=Samarth-MacBook-Pro.local
ARG GIT_COMMIT_AUTHOR=samarth-sh


# Run tests (optional but recommended)
RUN GOOS=linux CGO_ENABLED=1 go test -v -a -tags musl

# Build the Go app with appropriate build flags for versioning
RUN GOOS=linux CGO_ENABLED=1 go build -a -tags musl -o /go/bin/go-backend \
  -ldflags "-X github.com/your_username/go-actuator.BuildStamp=$BUILD_STAMP \
            -X github.com/your_username/go-actuator.GitCommitID=$GIT_COMMIT_ID \
            -X github.com/your_username/go-actuator.GitPrimaryBranch=$GIT_PRIMARY_BRANCH \
            -X github.com/your_username/go-actuator.GitURL=$GIT_URL \
            -X github.com/your_username/go-actuator.Username=$HOST_NAME \
            -X github.com/your_username/go-actuator.core.GitCommitAuthor=$GIT_COMMIT_AUTHOR"

##
# Archive Stage
##

# Base image for final build
FROM 732165046977.dkr.ecr.ap-south-1.amazonaws.com/sre-golang-base-image:3  AS artifact

# Set the service port
ARG SERVICE_PORT=8080

# Set the timezone
ENV TZ Asia/Calcutta

# Copy the compiled binary from the build stage
COPY --from=builder /go/bin/go-backend /app/go-backend

# Copy MSSQL stored procedures and user-defined functions
COPY MSSQL-SP /app/MSSQL-SP
COPY MSSQL-UDF /app/MSSQL-UDF

# Healthcheck and expose port
HEALTHCHECK NONE
EXPOSE ${SERVICE_PORT}

# Create an entrypoint script
RUN echo -e "#!/bin/sh\n/app/go-backend" > /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Set the entrypoint to run the Go app
ENTRYPOINT ["/app/entrypoint.sh"]

