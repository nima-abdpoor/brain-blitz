# Use the official Golang image as a base
FROM golang:1.20 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy
RUN go mod download

# Copy the source code into the container
COPY . .

# Install sql-migrate
RUN go install github.com/rubenv/sql-migrate/...@latest

# Build the Go app
RUN go build -o main ./delivery/http

# Start a new stage from scratch using the same base image
FROM golang:1.20

# Start a new stage from scratch
FROM ubuntu:22.04

# Install necessary libraries (glibc)
RUN apt-get update && apt-get install -y libc6 && apt-get install -y netcat && rm -rf /var/lib/apt/lists/*

# Set the Current Working Directory inside the container
WORKDIR /app



# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /go/bin/sql-migrate /usr/local/bin/sql-migrate
COPY config.yml .
COPY buf.gen.yaml .
COPY buf.yaml .

# Copy migration files and configuration
COPY internal/infra/config/db/dbconfig.yml ./internal/infra/config/db/
COPY internal/infra/repository/migration ./internal/infra/repository/migration

# Expose port 8080 to the outside world
EXPOSE 8000

# Command to run the executable
CMD ["sh", "-c", "sleep 30 && sql-migrate up -env=production -config=internal/infra/config/db/dbconfig.yml && ./main"]
