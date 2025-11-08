package mcp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	cfgpkg "github.com/tuusuario/afs-challenge/internal/config"
	_ "github.com/lib/pq"
)

// MCPClient provides access to Tiger Cloud forks via direct PostgreSQL connections.
// Uses pre-created, immutable forks (gwb579t287, mn4o89xewb) with zero-copy PITR support.
type MCPClient struct {
	configDir  string
	publicKey  string
	secretKey  string
	projectID  string
	fork1URL   string
	fork2URL   string
	mainURL    string
	timeout    time.Duration
	maxRetries int
}

// QueryResult is a normalized subset of query results.
type QueryResult struct {
	Rows            []map[string]any `json:"rows,omitempty"`
	RowCount        int              `json:"row_count,omitempty"`
	ExecutionTimeMs float64          `json:"execution_time_ms,omitempty"`
	Command         string           `json:"command,omitempty"`
	Message         string           `json:"message,omitempty"`
}

// New creates a new MCP client using direct PostgreSQL connections.
func New(c *cfgpkg.Config, httpClient interface{}) (*MCPClient, error) {
	if c == nil {
		return nil, errors.New("nil config")
	}
	
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			configDir = "/root/.config/tiger"
		} else {
			configDir = home + "/.config/tiger"
		}
	}
	
	// Pre-load fork connection strings from environment
	fork1URL := os.Getenv("TIGER_FORK_A1_SERVICE_URL")
	fork2URL := os.Getenv("TIGER_FORK_A2_SERVICE_URL")
	mainURL := os.Getenv("TIGER_MAIN_SERVICE_URL")
	
	cl := &MCPClient{
		configDir:  configDir,
		publicKey:  c.TigerCloud.PublicKey,
		secretKey:  c.TigerCloud.SecretKey,
		projectID:  c.TigerCloud.ProjectID,
		fork1URL:   fork1URL,
		fork2URL:   fork2URL,
		mainURL:    mainURL,
		timeout:    30 * time.Second,
		maxRetries: 3,
	}
	return cl, nil
}

// Connect verifies tiger CLI is available and authenticated.
// If credentials are provided, it runs tiger auth login to store them.
// In production (e.g., Railway), tiger CLI may not be available - skip MCP auth.
func (c *MCPClient) Connect(ctx context.Context) error {
	// Check if tiger CLI is available
	statusArgs := []string{"--version"}
	statusCmd := exec.CommandContext(ctx, "tiger", statusArgs...)
	if err := statusCmd.Run(); err != nil {
		fmt.Printf("[Tiger Auth] ⚠️  Tiger CLI not available - skipping MCP auth (using direct PostgreSQL)\n")
		return nil // Don't fail, just use direct DB connections
	}

	// If credentials provided, login first
	if c.publicKey != "" && c.secretKey != "" && c.projectID != "" {
		fmt.Printf("[Tiger Auth] Attempting login with credentials...\n")
		loginArgs := []string{
			"--config-dir", c.configDir,
			"auth", "login",
			"--public-key", c.publicKey,
			"--secret-key", c.secretKey,
			"--project-id", c.projectID,
		}
		loginCmd := exec.CommandContext(ctx, "tiger", loginArgs...)
		if output, err := loginCmd.CombinedOutput(); err != nil {
			fmt.Printf("[Tiger Auth] ⚠️  Login failed, continuing with direct PostgreSQL: %v\n", err)
			return nil // Don't fail
		}
		fmt.Printf("[Tiger Auth] ✅ Login successful\n")
	}
	
	// Check status after login
	fmt.Printf("[Tiger Auth] Verifying authentication status...\n")
	statusCmd = exec.CommandContext(ctx, "tiger", []string{"--config-dir", c.configDir, "auth", "status"}...)
	
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Tiger Auth] Status check failed (non-fatal): %v\n", err)
		return nil
	}
	fmt.Printf("[Tiger Auth] Status: %s\n", string(output))
	return nil
}

// ExecuteQuery runs a SQL query via direct PostgreSQL connection.
func (c *MCPClient) ExecuteQuery(ctx context.Context, serviceID, sqlQuery string, timeoutMs int) (QueryResult, error) {
	fmt.Printf("      [ExecuteQuery] Using PostgreSQL direct connection\n")
	return c.executeQueryPostgres(ctx, serviceID, sqlQuery, timeoutMs)
}

// executeQueryPostgres executes query via direct PostgreSQL connection.
func (c *MCPClient) executeQueryPostgres(ctx context.Context, serviceID, sqlQuery string, timeoutMs int) (QueryResult, error) {
	// Get connection string for pre-created fork or main service
	var connStr string
	switch serviceID {
	case "gwb579t287": // TIGER_FORK_A1_SERVICE_ID
		connStr = c.fork1URL
	case "mn4o89xewb": // TIGER_FORK_A2_SERVICE_ID
		connStr = c.fork2URL
	case "wuj5xa6zpz": // TIGER_MAIN_SERVICE_ID
		connStr = c.mainURL
	default:
		// Try to match by name
		if strings.Contains(serviceID, "agent-1") {
			connStr = c.fork1URL
		} else if strings.Contains(serviceID, "agent-2") {
			connStr = c.fork2URL
		} else if strings.Contains(serviceID, "main") {
			connStr = c.mainURL
		} else {
			return QueryResult{}, fmt.Errorf("mcp: unknown fork service ID: %s", serviceID)
		}
	}
	
	if connStr == "" {
		return QueryResult{}, fmt.Errorf("mcp: connection string not configured for fork %s", serviceID)
	}
	
	// Set timeout
	queryCtx := ctx
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	
	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return QueryResult{}, fmt.Errorf("mcp: failed to open database: %w", err)
	}
	defer db.Close()
	
	// Test connection
	if err := db.PingContext(queryCtx); err != nil {
		return QueryResult{}, fmt.Errorf("mcp: failed to connect to database: %w", err)
	}
	
	// Execute query
	rows, err := db.QueryContext(queryCtx, sqlQuery)
	if err != nil {
		return QueryResult{}, fmt.Errorf("mcp: query execution failed: %w", err)
	}
	defer rows.Close()
	
	// Get columns
	columns, err := rows.Columns()
	if err != nil {
		return QueryResult{}, fmt.Errorf("mcp: failed to get columns: %w", err)
	}
	
	// Fetch rows
	result := QueryResult{Rows: []map[string]any{}}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return QueryResult{}, fmt.Errorf("mcp: failed to scan row: %w", err)
		}
		
		entry := make(map[string]any)
		for i, col := range columns {
			entry[col] = values[i]
		}
		result.Rows = append(result.Rows, entry)
	}
	
	if err := rows.Err(); err != nil {
		return QueryResult{}, fmt.Errorf("mcp: error reading rows: %w", err)
	}
	
	result.RowCount = len(result.Rows)
	return result, nil
}

// CreateFork returns a pre-created fixed fork ID based on fork name.
// Forks are permanent (never deleted), data is immutable.
func (c *MCPClient) CreateFork(ctx context.Context, parentServiceID, forkName string) (string, error) {
	var forkID string
	if strings.Contains(forkName, "agent-2") || strings.Contains(forkName, "agent_2") {
		forkID = "mn4o89xewb" // TIGER_FORK_A2_SERVICE_ID
	} else {
		forkID = "gwb579t287" // TIGER_FORK_A1_SERVICE_ID
	}
	
	fmt.Printf("      ✅ Using pre-created fork: %s\n", forkID)
	return forkID, nil
}

// DeleteFork is a no-op since forks are pre-created and permanent.
func (c *MCPClient) DeleteFork(ctx context.Context, serviceID string) error {
	fmt.Printf("      [DeleteFork] No-op: forks are permanent\n")
	return nil
}

// Close is a no-op for direct PostgreSQL connections.
func (c *MCPClient) Close() error { return nil }
