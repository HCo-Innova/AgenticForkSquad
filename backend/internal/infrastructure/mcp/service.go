package mcp

// ServiceInfo represents minimal metadata for a Tiger service/fork.
type ServiceInfo struct {
	ServiceID    string   `json:"service_id"`
	ParentID     string   `json:"parent_id,omitempty"`
	Status       string   `json:"status,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	SizeMB       float64  `json:"size_mb,omitempty"`
	LastAccessed string   `json:"last_accessed,omitempty"`
	Tables       []string `json:"tables,omitempty"`
}

// Legacy HTTP methods (kept for reference, all delegated to client.go CLI proxy methods)
// These are now handled by the MCPClient CLI proxy implementation in client.go
// CreateFork, CreateForkAtTimestamp, DeleteFork, ExecuteQuery are implemented in client.go
