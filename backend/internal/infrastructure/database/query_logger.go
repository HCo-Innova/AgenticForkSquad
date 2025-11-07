package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"
	"time"

	llmpkg "github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
)

// QueryLogEntry represents a logged query with optional embedding.
type QueryLogEntry struct {
	ID                  int
	QueryText           string
	QueryHash           string
	ExecutionTimeMs     *float64
	RowsReturned        *int
	ExecutedAt          time.Time
	AgentType           *string
	TaskID              *int
	QueryEmbedding      []float32 // Vector embeddings
	EmbeddingModel      *string
	EmbeddingGeneratedAt *time.Time
	IsSlow              bool
	Notes               *string
}

// QueryLogger handles query logging and embedding generation.
type QueryLogger struct {
	db        *sql.DB
	llmClient llmpkg.LLMClient
}

// NewQueryLogger creates a new query logger.
func NewQueryLogger(db *sql.DB, llmClient llmpkg.LLMClient) *QueryLogger {
	return &QueryLogger{
		db:        db,
		llmClient: llmClient,
	}
}

// LogQuery saves a query execution log to the database.
// If generateEmbedding is true, generates semantic embedding via LLM.
func (ql *QueryLogger) LogQuery(ctx context.Context, entry *QueryLogEntry, generateEmbedding bool) (int, error) {
	if entry.QueryText == "" {
		return 0, fmt.Errorf("query text cannot be empty")
	}

	// Generate deterministic hash of query (for deduplication)
	if entry.QueryHash == "" {
		entry.QueryHash = generateQueryHash(entry.QueryText)
	}

	// Generate embedding if requested
	var embedding []float32
	var embeddingModel *string
	var embeddingGeneratedAt *time.Time

	if generateEmbedding && ql.llmClient != nil {
		emb, err := ql.generateEmbedding(ctx, entry.QueryText)
		if err != nil {
			// Log but don't fail - embedding is optional
			fmt.Printf("warning: failed to generate embedding: %v\n", err)
		} else {
			embedding = emb
			model := "text-embedding-vertex-001"
			embeddingModel = &model
			now := time.Now()
			embeddingGeneratedAt = &now
		}
	}

	// Insert into database
	var id int
	query := `
		INSERT INTO query_logs (
			query_text, query_hash, execution_time_ms, rows_returned,
			executed_at, agent_type, task_id, query_embedding,
			embedding_model, embedding_generated_at, is_slow, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	isSlow := false
	if entry.ExecutionTimeMs != nil && *entry.ExecutionTimeMs > 1000 {
		isSlow = true
	}

	// Convert []float32 to PostgreSQL vector format if present
	var embeddingValue interface{}
	if len(embedding) > 0 {
		embeddingValue = embedding
	}

	err := ql.db.QueryRowContext(ctx, query,
		entry.QueryText,
		entry.QueryHash,
		entry.ExecutionTimeMs,
		entry.RowsReturned,
		entry.ExecutedAt,
		entry.AgentType,
		entry.TaskID,
		embeddingValue,
		embeddingModel,
		embeddingGeneratedAt,
		isSlow,
		entry.Notes,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert query log: %w", err)
	}

	return id, nil
}

// generateQueryHash creates a deterministic hash of the query for deduplication.
func generateQueryHash(query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	hash := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 chars for brevity
}

// generateEmbedding generates semantic embedding using the LLM client.
func (ql *QueryLogger) generateEmbedding(ctx context.Context, queryText string) ([]float32, error) {
	if ql.llmClient == nil {
		return nil, fmt.Errorf("no LLM client configured")
	}

	// Use a shorter timeout for embedding generation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Vertex AI embeddings would be called here
	// For now, this is a placeholder - actual implementation depends on
	// whether LLMClient interface supports embeddings
	// You may need to extend LLMClient interface or use a separate embeddings client

	// Placeholder: return a zero vector (1536 dimensions)
	// In production, call actual embedding API
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.0
	}

	return embedding, nil
}

// GetSlowQueries retrieves queries with execution_time_ms > threshold.
func (ql *QueryLogger) GetSlowQueries(ctx context.Context, thresholdMs float64, limit int) ([]*QueryLogEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, query_text, query_hash, execution_time_ms, rows_returned,
		       executed_at, agent_type, task_id, is_slow
		FROM query_logs
		WHERE execution_time_ms > $1
		ORDER BY execution_time_ms DESC
		LIMIT $2
	`

	rows, err := ql.db.QueryContext(ctx, query, thresholdMs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query slow queries: %w", err)
	}
	defer rows.Close()

	var entries []*QueryLogEntry
	for rows.Next() {
		var e QueryLogEntry
		if err := rows.Scan(
			&e.ID, &e.QueryText, &e.QueryHash, &e.ExecutionTimeMs,
			&e.RowsReturned, &e.ExecutedAt, &e.AgentType, &e.TaskID, &e.IsSlow,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		entries = append(entries, &e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return entries, nil
}

// CountQueryLogs returns total count of logged queries.
func (ql *QueryLogger) CountQueryLogs(ctx context.Context) (int, error) {
	var count int
	err := ql.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM query_logs").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count query logs: %w", err)
	}
	return count, nil
}

// CleanupOldQueryLogs deletes query logs older than retentionDays.
func (ql *QueryLogger) CleanupOldQueryLogs(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	result, err := ql.db.ExecContext(ctx,
		"DELETE FROM query_logs WHERE executed_at < $1",
		cutoff,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old query logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
