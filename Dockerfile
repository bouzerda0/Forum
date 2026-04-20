# Build stage
FROM golang:1.22-alpine AS builder

# Install gcc and libc-dev for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -o forum .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install tzdata and ca-certificates
RUN apk add --no-cache tzdata ca-certificates

# Copy the binary from builder
COPY --from=builder /app/forum .

# Copy templates and static assets
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Ensure the database directory exists and has correct permissions
RUN mkdir -p /app/database && chmod 777 /app/database

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./forum"]
