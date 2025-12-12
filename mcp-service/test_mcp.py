#!/usr/bin/env python3
"""
Test script for MCP server
This tests the MCP protocol implementation by sending JSON-RPC 2.0 requests
"""

import json
import subprocess
import time
import sys

def test_mcp_server():
    print("Starting MCP Server Test...")
    print("=" * 50)
    
    # Start the MCP server
    try:
        process = subprocess.Popen(
            ['./mcp-server'],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1,
            env={
                'DATABASE_URL': 'host=localhost port=5432 user=postgres password=postgres dbname=mcp_restaurant sslmode=disable'
            }
        )
    except FileNotFoundError:
        print("Error: mcp-server binary not found. Run 'go build -o mcp-server cmd/mcp/main.go' first")
        return
    
    time.sleep(0.5)  # Give server time to start
    
    def send_request(request, description):
        print(f"\n{description}")
        print(f"Request: {json.dumps(request)}")
        try:
            process.stdin.write(json.dumps(request) + '\n')
            process.stdin.flush()
            time.sleep(0.3)
            
            # Try to read response (non-blocking)
            # Note: This is simplified; real implementation would need better handling
            return True
        except Exception as e:
            print(f"Error: {e}")
            return False
    
    # Test 1: Initialize
    send_request({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "test-client",
                "version": "1.0.0"
            }
        }
    }, "Test 1: Initialize")
    
    # Test 2: Notifications/Initialized
    send_request({
        "jsonrpc": "2.0",
        "method": "notifications/initialized"
    }, "Test 2: Notifications/Initialized")
    
    # Test 3: List Tools
    send_request({
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list"
    }, "Test 3: List Tools")
    
    # Test 4: Call get_restaurants tool
    send_request({
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {
            "name": "get_restaurants",
            "arguments": {}
        }
    }, "Test 4: Call get_restaurants tool")
    
    # Test 5: Call get_menu tool
    send_request({
        "jsonrpc": "2.0",
        "id": 4,
        "method": "tools/call",
        "params": {
            "name": "get_menu",
            "arguments": {
                "restaurant_id": 1
            }
        }
    }, "Test 5: Call get_menu tool for restaurant 1")
    
    # Test 6: Invalid method
    send_request({
        "jsonrpc": "2.0",
        "id": 5,
        "method": "invalid_method"
    }, "Test 6: Invalid method (should return error)")
    
    print("\n" + "=" * 50)
    print("Waiting for responses...")
    time.sleep(1)
    
    # Collect all output
    print("\n" + "=" * 50)
    print("Server Responses:")
    print("=" * 50)
    
    # Close stdin and wait for output
    try:
        process.stdin.close()
    except:
        pass
    
    try:
        stdout, stderr = process.communicate(timeout=2)
    except subprocess.TimeoutExpired:
        process.kill()
        stdout, stderr = process.communicate()
        print("\nServer timed out")
    except Exception as e:
        print(f"\nError communicating with server: {e}")
        process.kill()
        try:
            stdout, stderr = process.communicate(timeout=1)
        except:
            stdout = stderr = ""
    
    if stdout:
        print("\nSTDOUT (JSON-RPC responses):")
        for line in stdout.strip().split('\n'):
            if line:
                try:
                    response = json.loads(line)
                    print(json.dumps(response, indent=2))
                except json.JSONDecodeError:
                    print(line)
    
    if stderr:
        print("\nSTDERR (Server logs):")
        print(stderr)
    
    print("\n" + "=" * 50)
    print("Test complete!")
    
    # Check if server had database connection issues
    if stderr and "Failed to connect to database" in stderr:
        print("\nNote: Database connection failed. This is expected if PostgreSQL is not running.")
        print("To fully test with database, ensure PostgreSQL is running:")
        print("  docker-compose up -d")
        print("  or")
        print("  sudo systemctl start postgresql")
        return 1
    
    return 0

if __name__ == "__main__":
    sys.exit(test_mcp_server())
