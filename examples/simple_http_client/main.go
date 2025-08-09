package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// MCPRequest represents a JSON-RPC MCP request
type MCPRequest struct {
	JSONRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id"`
}

// MCPResponse represents a JSON-RPC MCP response
type MCPResponse struct {
	JSONRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      int         `json:"id"`
}

func main() {
	// Get API key from environment or use default
	apiKey := os.Getenv("MCP_API_KEY")
	if apiKey == "" {
		apiKey = "mcpfier_dev_123456"
		fmt.Printf("Using default API key: %s\n", apiKey)
	}

	serverURL := "http://localhost:8080"
	if len(os.Args) > 1 {
		serverURL = os.Args[1]
	}

	fmt.Printf("Testing MCPFier HTTP server at %s\n", serverURL)

	// Test health endpoint
	fmt.Println("\n1. Testing health endpoint...")
	testHealthEndpoint(serverURL)

	// Test MCP endpoint without auth (should fail)
	fmt.Println("\n2. Testing MCP endpoint without auth (should fail)...")
	testMCPEndpoint(serverURL, "", "tools/list", nil)

	// Test MCP endpoint with invalid API key (should fail)
	fmt.Println("\n3. Testing MCP endpoint with invalid API key (should fail)...")
	testMCPEndpoint(serverURL, "invalid_key", "tools/list", nil)

	// Test MCP endpoint with valid API key
	fmt.Println("\n4. Testing MCP endpoint with valid API key...")
	testMCPEndpoint(serverURL, apiKey, "tools/list", nil)

	// Test calling a specific tool
	fmt.Println("\n5. Testing echo-test tool...")
	testMCPEndpoint(serverURL, apiKey, "tools/call", map[string]interface{}{
		"name":      "echo-test",
		"arguments": map[string]interface{}{},
	})
}

func testHealthEndpoint(serverURL string) {
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Health check: %d %s\n", resp.StatusCode, string(body))
}

func testMCPEndpoint(serverURL, apiKey, method string, params interface{}) {
	request := MCPRequest{
		JSONRpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", serverURL+"/mcp", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode == 200 {
		var mcpResp MCPResponse
		if err := json.Unmarshal(body, &mcpResp); err == nil {
			fmt.Printf("✅ Success: %s\n", formatResponse(&mcpResp))
		} else {
			fmt.Printf("✅ Success (raw): %s\n", string(body))
		}
	} else {
		fmt.Printf("❌ Failed (%d): %s\n", resp.StatusCode, string(body))
	}
}

func formatResponse(resp *MCPResponse) string {
	if resp.Error != nil {
		return fmt.Sprintf("Error: %v", resp.Error)
	}
	
	if result, ok := resp.Result.(map[string]interface{}); ok {
		if tools, exists := result["tools"]; exists {
			return fmt.Sprintf("Found %d tools", len(tools.([]interface{})))
		}
		if content, exists := result["content"]; exists {
			if contentArray, ok := content.([]interface{}); ok && len(contentArray) > 0 {
				if textContent, ok := contentArray[0].(map[string]interface{}); ok {
					return fmt.Sprintf("Tool output: %v", textContent["text"])
				}
			}
		}
	}
	
	return fmt.Sprintf("Result: %v", resp.Result)
}