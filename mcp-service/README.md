# MCP Service - Restaurant Management System

This is a Microservice Control Plane (MCP) service for managing Indian restaurants, featuring order management, billing, and inventory tracking.

## Features

- Restaurant management (CRUD operations)
- Indian cuisine-focused menu management with categories like North Indian, South Indian, Street Food
- Dietary preference tagging (Vegetarian, Non-Vegetarian, Vegan, Jain-friendly)
- Spice level indicators for dishes
- Order processing system with billing capabilities
- Tax calculation (GST) and discount support
- Multiple payment method support (Cash, Card, UPI, Digital Wallet)
- PostgreSQL database integration with sample Indian food data

## Architecture

- **Backend**: Go with Gorilla Mux router
- **Database**: PostgreSQL for persistent storage
- **Containerization**: Docker
- **Orchestration**: Kubernetes with StatefulSets for database

## API Endpoints

All endpoints are prefixed with `/mcp`:

### Restaurants
- `GET    /mcp/restaurants` - Get all restaurants
- `POST   /mcp/restaurants` - Create a new restaurant
- `GET    /mcp/restaurants/{id}` - Get a specific restaurant
- `PUT    /mcp/restaurants/{id}` - Update a restaurant
- `DELETE /mcp/restaurants/{id}` - Delete a restaurant

### Menu Items
- `GET    /mcp/restaurants/{id}/menu` - Get menu items for a restaurant
- `POST   /mcp/restaurants/{id}/menu` - Add a new menu item

### Orders
- `GET    /mcp/orders` - Get all orders
- `POST   /mcp/orders` - Create a new order
- `GET    /mcp/orders/{id}` - Get a specific order

### Health Check
- `GET    /health` - Health check endpoint

## Local Development

### Prerequisites

- Go 1.21+
- PostgreSQL

### Setup

1. Set up PostgreSQL with the connection string:
   ```bash
   export DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
   ```

2. Run database migrations (the application will create tables automatically)

3. Start the service:
   ```bash
   go run cmd/api/main.go
   ```

## Kubernetes Deployment

The service comes with Kubernetes manifests for easy deployment:

1. Navigate to the k8s directory
2. Apply the database manifest:
   ```bash
   kubectl apply -f database.yaml
   ```
3. Apply the application manifest:
   ```bash
   kubectl apply -f deployment.yaml
   ```

## Sample Data

On startup, the service automatically seeds the database with:

- 3 Indian restaurants (Taj Mahal, Surya Mahal, Hyderabad House)
- 18 Indian menu items covering various regional cuisines and dietary preferences