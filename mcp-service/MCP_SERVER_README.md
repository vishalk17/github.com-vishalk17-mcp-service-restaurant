# MCP Server - Restaurant Management

A properly implemented **Model Context Protocol (MCP)** server for managing Indian restaurants, featuring order management, billing, and inventory tracking.

## What Was Fixed

The original implementation was **incorrectly designed** as an HTTP/WebSocket server, which is NOT how MCP servers work. Here's what was wrong and what was fixed:

### Problems in Original Implementation

1. **Wrong Transport Layer**: Used WebSocket over HTTP instead of stdio
2. **Incorrect Protocol**: Missing proper MCP protocol structure (serverInfo, proper capabilities)
3. **Buggy Tool Calls**: Variable name bugs (line 235 used `id` instead of `args["restaurant_id"]`)
4. **Missing JSON Schema**: Tools didn't have proper `inputSchema` definitions
5. **Wrong Architecture**: MCP servers use stdio for communication, not HTTP endpoints

### What Was Fixed

1. ✅ **Proper stdio Transport**: MCP server now uses stdin/stdout for JSON-RPC 2.0 communication
2. ✅ **Correct MCP Protocol**: Implements proper `initialize`, `notifications/initialized`, `tools/list`, and `tools/call` methods
3. ✅ **Complete JSON Schema**: All tools have proper `inputSchema` with types, descriptions, and required fields
4. ✅ **Fixed All Bugs**: Variable scoping and type conversion issues resolved
5. ✅ **MCP Specification Compliance**: Follows the official MCP specification (protocol version 2024-11-05)

## Architecture

- **Transport**: stdio (standard input/output)
- **Protocol**: JSON-RPC 2.0 over stdio
- **Database**: PostgreSQL
- **Language**: Go 1.21+

## Available Tools

The MCP server exposes the following tools:

### 1. `get_restaurants`
Get a list of all Indian restaurants with their details.

**Parameters**: None

### 2. `get_restaurant`
Get details of a specific restaurant by ID.

**Parameters**:
- `restaurant_id` (integer, required): The ID of the restaurant to retrieve

### 3. `get_menu`
Get the menu items for a specific restaurant, including Indian dishes with dietary preferences and spice levels.

**Parameters**:
- `restaurant_id` (integer, required): The ID of the restaurant whose menu to retrieve

### 4. `create_restaurant`
Create a new restaurant with details.

**Parameters**:
- `name` (string, required): Name of the restaurant
- `address` (string, required): Address of the restaurant
- `phone_number` (string, optional): Phone number of the restaurant
- `cuisine_type` (string, optional): Type of cuisine (defaults to "Indian")

### 5. `get_orders`
Get a list of all orders with their details including customer info, items, billing, and payment status.

**Parameters**: None

### 6. `get_order`
Get details of a specific order by ID.

**Parameters**:
- `order_id` (integer, required): The ID of the order to retrieve

### 7. `create_order`
Create a new order with items, customer details, and payment information. GST tax (5%) will be automatically calculated.

**Parameters**:
- `restaurant_id` (integer, required): ID of the restaurant
- `customer_name` (string, required): Name of the customer
- `items` (array, required): Array of order items with menu_item_id, quantity, price, and optional notes
- `customer_phone` (string, optional): Phone number of the customer
- `discount` (number, optional): Discount amount (defaults to 0)
- `payment_method` (string, optional): Payment method (cash, card, upi, digital_wallet)
- `billing_address` (string, optional): Billing address

## Setup & Running

### Prerequisites

1. Go 1.21 or later
2. PostgreSQL database
3. Docker (optional, for containerized database)

### Build

```bash
cd /home/vishalk17/mcp-service/mcp-service
go build -o mcp-server cmd/mcp/main.go
```

### Start Database

Using Docker Compose:
```bash
docker-compose up -d postgres
```

Or use an existing PostgreSQL instance and set the `DATABASE_URL` environment variable.

### Run the Server

```bash
export DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
./mcp-server
```

The server will:
1. Connect to the database
2. Create tables if they don't exist
3. Seed sample data (3 restaurants with 18 Indian menu items)
4. Listen on stdin for JSON-RPC 2.0 requests

## Testing

### Manual Test (Simple)

Send JSON-RPC 2.0 requests via stdin:

```bash
# Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./mcp-server

# List tools
(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'; \
 echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; \
 echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}') | ./mcp-server

# Call get_restaurants tool
(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'; \
 echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; \
 echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_restaurants","arguments":{}}}') | ./mcp-server
```

### Using MCP Inspector

The MCP Inspector is a tool for testing MCP servers. Configure it with `mcp-config.json`:

```json
{
  "mcpServers": {
    "restaurant-server": {
      "command": "/home/vishalk17/mcp-service/mcp-service/mcp-server",
      "env": {
        "DATABASE_URL": "host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
      }
    }
  }
}
```

Then run:
```bash
npx @modelcontextprotocol/inspector mcp-config.json
```

## Configuration for AI Clients

To use this MCP server with Claude Desktop or other MCP clients, add to your configuration:

### Claude Desktop (macOS)
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "restaurant": {
      "command": "/home/vishalk17/mcp-service/mcp-service/mcp-server",
      "env": {
        "DATABASE_URL": "host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
      }
    }
  }
}
```

### Claude Desktop (Windows)
Edit `%APPDATA%\Claude\claude_desktop_config.json` with the same configuration (adjust path).

## Protocol Flow

1. **Initialize**: Client sends `initialize` request with protocol version and capabilities
2. **Server Response**: Server responds with its capabilities and server info
3. **Notification**: Client sends `notifications/initialized` to complete handshake
4. **List Tools**: Client calls `tools/list` to discover available tools
5. **Call Tools**: Client calls `tools/call` with tool name and arguments
6. **Results**: Server returns results in MCP format with content array

## Example Responses

### Initialize Response
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "restaurant-mcp-server",
      "version": "1.0.0"
    }
  }
}
```

### Tool Call Response (get_restaurants)
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "[{\"id\":1,\"name\":\"Taj Mahal Restaurant\",...}]"
      }
    ]
  }
}
```

## Logging

The server logs to **stderr** (stdout is reserved for JSON-RPC responses). Logs include:
- Database connection status
- Received requests
- Tool calls with arguments
- Errors and warnings

## Sample Data

On first run, the server seeds the database with:

**3 Restaurants:**
1. Taj Mahal Restaurant (New Delhi)
2. Surya Mahal (Mumbai)  
3. Hyderabad House (Hyderabad)

**18 Menu Items** covering:
- Main Course (Butter Chicken, Biryani, etc.)
- South Indian (Dosa, Idli, etc.)
- Street Food (Vada Pav, Pav Bhaji, etc.)
- Breads, Desserts, and Beverages

Each item includes dietary preferences (vegetarian/non-vegetarian) and spice levels (mild/medium/hot).

## Troubleshooting

### Database Connection Failed
- Ensure PostgreSQL is running: `docker-compose up -d postgres`
- Check DATABASE_URL environment variable
- Verify network connectivity to database

### No Response from Server
- Check that you're sending valid JSON-RPC 2.0 requests
- Ensure `jsonrpc: "2.0"` is included in every request
- Verify stdin is properly connected

### Tool Not Found
- Call `tools/list` first to see available tools
- Ensure server is initialized before calling tools

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [MCP SDK Documentation](https://github.com/modelcontextprotocol)
