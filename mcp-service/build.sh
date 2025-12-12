#!/bin/bash

set -e

echo "Building MCP Service..."

# Navigate to the project directory
cd /home/vishalk17/mcp-service/mcp-service

# Download dependencies
go mod tidy

# Install the application
go install ./cmd/api

echo "Build completed successfully!"
echo "To run the service locally, make sure PostgreSQL is running and execute:"
echo "DATABASE_URL='host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable' go run cmd/api/main.go"