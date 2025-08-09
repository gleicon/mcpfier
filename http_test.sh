#!/bin/bash

  echo "Testing MCPFier HTTP Server..."

  # Health check
  echo "1. Health check:"
  curl -s http://localhost:8080/health | jq .

  # Analytics dashboard (just check status)
  echo -e "\n2. Analytics dashboard:"
  curl -s -o /dev/null -w "Status: %{http_code}\n" http://localhost:8080/mcpfier/analytics

  # Test without auth (should fail)
  echo -e "\n3. MCP without auth (should fail):"
  curl -s http://localhost:8080/ -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"initialize","id":1}'

  # Initialize MCP session with auth
  echo -e "\n4. Initialize MCP session:"
  INIT_RESPONSE=$(curl -s http://localhost:8080/ \
    -H "Content-Type: application/json" \
    -H "X-API-Key: mcpfier_dev_123456" \
    -d '{
      "jsonrpc": "2.0",
      "method": "initialize",
      "id": 1,
      "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {},
        "clientInfo": {
          "name": "test-client",
          "version": "1.0"
        }
      }
    }')

  echo "$INIT_RESPONSE" | jq .

  # Extract session ID from headers (if available)
  SESSION_ID=$(curl -s -I http://localhost:8080/ \
    -H "Content-Type: application/json" \
    -H "X-API-Key: mcpfier_dev_123456" \
    -d '{
      "jsonrpc": "2.0",
      "method": "initialize",
      "id": 1,
      "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {},
        "clientInfo": {"name": "test-client", "version": "1.0"}
      }
    }' | grep -i "X-Session-ID" | cut -d' ' -f2 | tr -d '\r')

  echo "Session ID: $SESSION_ID"

  # Call echo-test tool
  echo -e "\n5. Call echo-test tool:"
  curl -s http://localhost:8080/ \
    -H "Content-Type: application/json" \
    -H "X-API-Key: mcpfier_dev_123456" \
    -H "X-Session-ID: $SESSION_ID" \
    -d '{
      "jsonrpc": "2.0",
      "method": "tools/call",
      "id": 2,
      "params": {
        "name": "echo-test",
        "arguments": {}
      }
    }' | jq .

  echo -e "\n\nServer logs should show all these requests with proper authentication!"

