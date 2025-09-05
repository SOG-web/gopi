# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go modules)
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/api/main.go -o ./docs

# Build the application
RUN go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy the docs directory for Swagger
COPY --from=builder /app/docs ./docs

# Create directories for logs and uploads
RUN mkdir -p /app/logs /app/uploads && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run
CMD ["./main"]
