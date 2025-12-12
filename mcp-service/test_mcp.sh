#!/bin/bash

# Test script for MCP server
# This sends JSON-RPC 2.0 requests to the MCP server via stdin

echo "Starting MCP Server Test..."
echo "==========================="

# Set database URL
export DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable"

# Create a named pipe for communication
PIPE=$(mktemp -u)
mkfifo "$PIPE"

# Start the MCP server in background
./mcp-server < "$PIPE" > test_output.log 2> test_errors.log &
SERVER_PID=$!

# Give server time to start
sleep 1

# Function to send request and wait for response
send_request() {
    local request="$1"
    echo "Sending: $request"
    echo "$request" > "$PIPE"
    sleep 0.5
}

echo ""
echo "Test 1: Initialize"
send_request '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'

echo ""
echo "Test 2: Notifications/Initialized"
send_request '{"jsonrpc":"2.0","method":"notifications/initialized"}'

echo ""
echo "Test 3: List Tools"
send_request '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'

echo ""
echo "Test 4: Call get_restaurants tool"
send_request '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_restaurants","arguments":{}}}'

echo ""
echo "Test 5: Call get_menu tool for restaurant 1"
send_request '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_menu","arguments":{"restaurant_id":1}}}'

echo ""
echo "Waiting for server to process requests..."
sleep 2

# Clean up
kill $SERVER_PID 2>/dev/null
rm -f "$PIPE"

echo ""
echo "==========================="
echo "Test complete! Check test_output.log and test_errors.log for results"
echo ""
echo "Server output:"
cat test_output.log
echo ""
echo "Server errors/logs:"
cat test_errors.log
