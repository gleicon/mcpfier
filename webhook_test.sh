#!/bin/bash

echo "Testing MCPFier MCP-to-API Gateway Functionality..."

# Initialize MCP session with auth
echo -e "\n1. Initialize MCP session:"
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

# Test 1: Simple GET request to httpbin
echo -e "\n2. Test webhook GET request (httpbin-get):"
curl -s http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcpfier_dev_123456" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "id": 2,
    "params": {
      "name": "httpbin-get",
      "arguments": {}
    }
  }' | jq .

# Test 2: POST request with JSON data
echo -e "\n3. Test webhook POST with JSON (httpbin-post-json):"
curl -s http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcpfier_dev_123456" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "id": 3,
    "params": {
      "name": "httpbin-post-json",
      "arguments": {}
    }
  }' | jq .

# Test 3: Basic authentication
echo -e "\n4. Test webhook with basic auth (webhook-basic-auth):"
curl -s http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcpfier_dev_123456" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "id": 4,
    "params": {
      "name": "webhook-basic-auth",
      "arguments": {}
    }
  }' | jq .

# Test 4: Compare with local command execution
echo -e "\n5. Test local command (echo-test) for comparison:"
curl -s http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mcpfier_dev_123456" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "id": 5,
    "params": {
      "name": "echo-test",
      "arguments": {}
    }
  }' | jq .

echo -e "\n\n=== Test Summary ==="
echo "-> Local execution: echo-test (execution_mode: local)"
echo "-> Webhook GET: httpbin-get (execution_mode: webhook)"
echo "-> Webhook POST: httpbin-post-json (execution_mode: webhook)"
echo "-> Webhook Auth: webhook-basic-auth (execution_mode: webhook)"
echo -e "\nServer analytics should show mixed execution modes!"
echo "Analytics dashboard: http://localhost:8080/mcpfier/analytics"
