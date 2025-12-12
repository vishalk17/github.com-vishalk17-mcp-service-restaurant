# Multi-stage build
FROM golang:1.21-alpine AS builder

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o mcp-api ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/mcp-api .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./mcp-api"]