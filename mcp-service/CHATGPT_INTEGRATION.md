# ChatGPT Integration Guide

## Important: ChatGPT Does NOT Support MCP

ChatGPT uses **OpenAPI/Actions**, not the Model Context Protocol (MCP). To use this with ChatGPT, you need to use the **HTTP API** instead of the MCP server.

## Setup for ChatGPT

### Step 1: Start the HTTP API Server

```bash
# Start the old HTTP API (not the MCP server)
cd /home/vishalk17/mcp-service/mcp-service
docker-compose up -d postgres
go run cmd/api/main.go
```

This starts the HTTP server on `http://localhost:8080`

### Step 2: Expose to Internet (Required for ChatGPT)

ChatGPT needs to access your API over the internet. Use a tunneling service:

#### Option A: ngrok
```bash
# Install ngrok
# https://ngrok.com/download

# Expose port 8080
ngrok http 8080
```

You'll get a URL like: `https://abc123.ngrok.io`

#### Option B: Cloudflare Tunnel
```bash
cloudflared tunnel --url http://localhost:8080
```

### Step 3: Create a GPT with Actions

1. Go to https://chat.openai.com
2. Click your profile → "My GPTs" → "Create a GPT"
3. Go to "Configure" tab
4. Scroll to "Actions" and click "Create new action"
5. Import the OpenAPI schema from `openapi.yaml`
6. Replace the server URL with your ngrok/cloudflare URL:

```yaml
servers:
  - url: https://your-ngrok-url.ngrok.io/api
```

### Step 4: Test

Ask your GPT:
- "List all restaurants"
- "Show me the menu for restaurant 1"
- "Create a new order for restaurant 1"

## Available Endpoints

All endpoints are prefixed with `/api`:

- `GET /api/restaurants` - List restaurants
- `GET /api/restaurants/{id}` - Get restaurant
- `GET /api/restaurants/{id}/menu` - Get menu
- `POST /api/restaurants` - Create restaurant
- `GET /api/orders` - List orders
- `GET /api/orders/{id}` - Get order
- `POST /api/orders` - Create order

## Summary: MCP vs ChatGPT

| Feature | MCP Server | ChatGPT Actions |
|---------|------------|-----------------|
| **Transport** | stdio | HTTP/HTTPS |
| **Protocol** | JSON-RPC 2.0 | REST API |
| **File to use** | `cmd/mcp/main.go` | `cmd/api/main.go` |
| **Schema** | MCP tools | OpenAPI spec |
| **Works with** | Claude Desktop | ChatGPT GPTs |
| **Requires internet** | No | Yes |
| **Port** | None (stdio) | 8080 |

## Recommendation

- **For Claude Desktop**: Use the MCP server (`mcp-server` binary)
- **For ChatGPT**: Use the HTTP API (`cmd/api/main.go`) with Actions
- **For Development**: Use both! They share the same database layer
