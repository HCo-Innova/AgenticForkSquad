package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPMCPClient provides direct HTTP communication with Tiger MCP server.
type HTTPMCPClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// MCPRequest represents a JSON-RPC 2.0 request to MCP server.
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      int64                  `json:"id"`
}

// MCPResponse represents a JSON-RPC 2.0 response from MCP server.
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
	ID      int64           `json:"id"`
}

// MCPError represents a JSON-RPC 2.0 error response.
type MCPError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// DBExecuteQueryParams represents parameters for db_execute_query method.
type DBExecuteQueryParams struct {
	ServiceID  string `json:"service_id"`
	Query      string `json:"query"`
	TimeoutMs  int    `json:"timeout_ms,omitempty"`
}

// DBExecuteQueryResult represents the result from db_execute_query.
type DBExecuteQueryResult struct {
	Rows       []map[string]interface{} `json:"rows,omitempty"`
	RowCount   int                      `json:"row_count,omitempty"`
	Message    string                   `json:"message,omitempty"`
	ExecutedAt time.Time                `json:"executed_at,omitempty"`
}

// NewHTTPMCPClient creates a new HTTP MCP client.
func NewHTTPMCPClient(mcpURL string, timeout time.Duration) *HTTPMCPClient {
	return &HTTPMCPClient{
		baseURL: mcpURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// ExecuteQuery executes a SQL query via HTTP MCP.
func (c *HTTPMCPClient) ExecuteQuery(ctx context.Context, serviceID, sqlQuery string, timeoutMs int) (*DBExecuteQueryResult, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		Method:  "db_execute_query",
		ID:      time.Now().UnixNano(),
		Params: map[string]interface{}{
			"service_id": serviceID,
			"query":      sqlQuery,
			"timeout_ms": timeoutMs,
		},
	}

	resp, err := c.call(ctx, &req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("mcp: db_execute_query failed: code=%d message=%s", resp.Error.Code, resp.Error.Message)
	}

	var result DBExecuteQueryResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("mcp: failed to parse query result: %w", err)
	}

	return &result, nil
}

// ServiceList lists all services via HTTP MCP.
func (c *HTTPMCPClient) ServiceList(ctx context.Context) ([]map[string]interface{}, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		Method:  "service_list",
		ID:      time.Now().UnixNano(),
		Params:  map[string]interface{}{},
	}

	resp, err := c.call(ctx, &req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("mcp: service_list failed: code=%d message=%s", resp.Error.Code, resp.Error.Message)
	}

	var services []map[string]interface{}
	if err := json.Unmarshal(resp.Result, &services); err != nil {
		return nil, fmt.Errorf("mcp: failed to parse services: %w", err)
	}

	return services, nil
}

// call sends a JSON-RPC request to the MCP server and returns the response.
func (c *HTTPMCPClient) call(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("mcp: failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("mcp: failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	// Execute request
	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("mcp: http request failed: %w (duration: %dms)", err, duration.Milliseconds())
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("mcp: failed to read response: %w", err)
	}

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mcp: unexpected HTTP status: %d (body: %s)", httpResp.StatusCode, string(respBody))
	}

	// Parse JSON-RPC response
	var mcpResp MCPResponse
	if err := json.Unmarshal(respBody, &mcpResp); err != nil {
		return nil, fmt.Errorf("mcp: failed to parse response: %w (body: %s)", err, string(respBody))
	}

	return &mcpResp, nil
}
