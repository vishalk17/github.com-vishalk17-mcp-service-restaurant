# Testing the MCP Service

## Prerequisites
- You need a running PostgreSQL instance to test the full application
- If PostgreSQL is not available, you can still verify compilation

## Compilation Test
The application compiles successfully if you see no errors:

```bash
cd /home/vishalk17/mcp-service/mcp-service
go build ./cmd/api
```

## Local PostgreSQL Setup for Full Testing

If you have PostgreSQL installed locally, you can:

1. Create the database:
```sql
CREATE DATABASE mcp_restaurant;
-- Connect to the database and create tables
```

2. Set environment variables:
```bash
export DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
export PORT="8080"
```

3. Run the application:
```bash
go run cmd/api/main.go
```

## Expected Output
Once running, you should see:
```
MCP Service starting on port 8080
Database connected successfully
```

## API Testing Examples
After starting the service, you can test with curl:

### Health Check
```bash
curl http://localhost:8080/health
```

### Get All Restaurants
```bash
curl http://localhost:8080/mcp/restaurants
```

### Create a Restaurant
```bash
curl -X POST http://localhost:8080/mcp/restaurants \
  -H "Content-Type: application/json" \
  -d '{"name":"New Restaurant","address":"123 Main St","phone_number":"+1234567890","cuisine_type":"Indian"}'
```

### Get Restaurant Menu
```bash
curl http://localhost:8080/mcp/restaurants/1/menu
```

## Docker Build Test
To verify Docker integration:
```bash
docker build -t mcp-service:latest .
```