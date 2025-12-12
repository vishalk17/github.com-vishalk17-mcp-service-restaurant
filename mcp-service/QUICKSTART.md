# Quick Start Guide

## What's Fixed?

The original code was **NOT a proper MCP server**. It used HTTP/WebSocket which is wrong. MCP servers use **stdio** (stdin/stdout) for JSON-RPC 2.0 communication.

## 30-Second Test

```bash
# 1. Start database (if not running)
docker-compose up -d postgres

# 2. Build the server
go build -o mcp-server cmd/mcp/main.go

# 3. Test it
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./mcp-server
```

You should see:
```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"restaurant-mcp-server","version":"1.0.0"}}}
```

## What's Available?

**7 Tools** for managing Indian restaurants:
- `get_restaurants` - List all restaurants
- `get_restaurant` - Get restaurant by ID
- `get_menu` - Get restaurant menu
- `create_restaurant` - Add new restaurant
- `get_orders` - List all orders
- `get_order` - Get order by ID  
- `create_order` - Create new order with GST

## Use with Claude Desktop

Edit `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "restaurant": {
      "command": "/full/path/to/mcp-server",
      "env": {
        "DATABASE_URL": "host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"
      }
    }
  }
}
```

Restart Claude Desktop. The tools will appear automatically.

## Key Differences: Old vs New

| Feature | Old (WRONG) | New (CORRECT) |
|---------|-------------|---------------|
| Transport | HTTP/WebSocket | stdio |
| Port | 8080 | None (uses stdin/stdout) |
| Protocol | Custom WebSocket | JSON-RPC 2.0 |
| Input Schema | Missing | Full JSON Schema |
| Server Info | Missing | Present |
| Tool Response | Direct data | MCP content format |

## Files

- ✅ `cmd/mcp/main.go` - **NEW**: Proper MCP server
- ❌ `cmd/api/main.go` - Old HTTP server (kept for compatibility)
- 📖 `MCP_SERVER_README.md` - Full documentation
- 📝 `FIXES_SUMMARY.md` - Detailed fixes
- 🧪 `test_mcp.py` - Test script
- ⚙️ `mcp-config.json` - MCP Inspector config

## Sample Commands

```bash
# List tools
(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'; \
 echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; \
 echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}') | ./mcp-server

# Get restaurants
(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'; \
 echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; \
 echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_restaurants","arguments":{}}}') | ./mcp-server
```

## Database

Sample data is automatically seeded:
- **3 restaurants** (Taj Mahal, Surya Mahal, Hyderabad House)
- **18 menu items** (Butter Chicken, Biryani, Dosa, etc.)
- Includes dietary info and spice levels

## Need Help?

- Read `MCP_SERVER_README.md` for detailed documentation
- Read `FIXES_SUMMARY.md` to understand what was fixed
- Check [MCP Specification](https://modelcontextprotocol.io/)
