package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	llmpkg "github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
)

// SimilarQuery represents a query result from hybrid search.
type SimilarQuery struct {
	ID              int
	QueryText       string
	ExecutionTimeMs *float64
	RowsReturned    *int
	ExecutedAt      time.Time
	TextScore       float64
	VectorScore     float64
	CombinedScore   float64
	Reason          string // Why it's similar (keyword/semantic)
}

// HybridSearchResult contains the full results from hybrid search.
type HybridSearchResult struct {
	InputQuery       string
	SimilarQueries   []*SimilarQuery
	SearchTimeMs     float64
	TextIndexUsed    bool
	VectorIndexUsed  bool
	TotalMatches     int
}

// HybridSearchService provides combined full-text and vector similarity search.
type HybridSearchService struct {
	db        *sql.DB
	llmClient llmpkg.LLMClient
}

// NewHybridSearchService creates a new hybrid search service.
func NewHybridSearchService(db *sql.DB, llmClient llmpkg.LLMClient) *HybridSearchService {
	return &HybridSearchService{
		db:        db,
		llmClient: llmClient,
	}
}

// HybridSearch performs combined full-text and vector similarity search.
// Text search has 40% weight, vector search has 60% weight.
// Returns top N most relevant similar queries.
func (hss *HybridSearchService) HybridSearch(ctx context.Context, inputQuery string, topN int) (*HybridSearchResult, error) {
	if inputQuery == "" {
		return nil, fmt.Errorf("input query cannot be empty")
	}

	if topN <= 0 {
		topN = 10
	}

	startTime := time.Now()

	// First, get results from full-text search
	textResults, err := hss.fullTextSearch(ctx, inputQuery, topN*2)
	if err != nil {
		return nil, fmt.Errorf("full-text search failed: %w", err)
	}

	// Second, get results from vector similarity search (if embeddings available)
	vectorResults := make(map[int]*SimilarQuery)
	if hss.llmClient != nil {
		// Generate embedding for input query
		embedding, err := hss.generateEmbedding(ctx, inputQuery)
		if err == nil && len(embedding) > 0 {
			vResults, err := hss.vectorSearch(ctx, embedding, topN*2)
			if err == nil {
				for _, q := range vResults {
					vectorResults[q.ID] = q
				}
			}
		}
	}

	// Combine results with weighted scoring
	combinedMap := make(map[int]*SimilarQuery)
	textIndexUsed := len(textResults) > 0
	vectorIndexUsed := len(vectorResults) > 0

	// Add text search results
	for _, q := range textResults {
		combinedMap[q.ID] = q
		combinedMap[q.ID].CombinedScore = q.TextScore * 0.4
		combinedMap[q.ID].Reason = "keyword match"
	}

	// Merge vector search results
	for id, q := range vectorResults {
		if existing, found := combinedMap[id]; found {
			// Already in text search - combine scores
			existing.VectorScore = q.VectorScore
			existing.CombinedScore = existing.TextScore*0.4 + q.VectorScore*0.6
			existing.Reason = "keyword + semantic"
		} else {
			// New entry from vector search
			q.CombinedScore = q.VectorScore * 0.6
			q.Reason = "semantic similarity"
			combinedMap[id] = q
		}
	}

	// Sort by combined score and take top N
	similar := make([]*SimilarQuery, 0, len(combinedMap))
	for _, q := range combinedMap {
		similar = append(similar, q)
	}

	// Sort by combined score (highest first)
	sortByScore(similar)

	if len(similar) > topN {
		similar = similar[:topN]
	}

	searchTime := time.Since(startTime).Milliseconds()

	return &HybridSearchResult{
		InputQuery:      inputQuery,
		SimilarQueries:  similar,
		SearchTimeMs:    float64(searchTime),
		TextIndexUsed:   textIndexUsed,
		VectorIndexUsed: vectorIndexUsed,
		TotalMatches:    len(combinedMap),
	}, nil
}

// fullTextSearch performs keyword-based search using PostgreSQL FTS.
func (hss *HybridSearchService) fullTextSearch(ctx context.Context, query string, limit int) ([]*SimilarQuery, error) {
	// Convert query to tsquery format for OR-based search
	// Split by spaces and join with | for OR matching
	terms := strings.Fields(strings.TrimSpace(query))
	if len(terms) == 0 {
		return []*SimilarQuery{}, nil
	}

	// Build tsquery string - OR all terms
	tsqueryStr := strings.Join(terms, "|")

	queryStr := `
		SELECT 
			id,
			query_text,
			execution_time_ms,
			rows_returned,
			executed_at,
			ts_rank(
				to_tsvector('english', query_text),
				to_tsquery('english', $1)
			) as text_score
		FROM query_logs
		WHERE to_tsvector('english', query_text) @@ to_tsquery('english', $1)
		  AND query_text != $2
		ORDER BY text_score DESC
		LIMIT $3
	`

	rows, err := hss.db.QueryContext(ctx, queryStr, tsqueryStr, query, limit)
	if err != nil {
		return nil, fmt.Errorf("full-text search query failed: %w", err)
	}
	defer rows.Close()

	var results []*SimilarQuery
	for rows.Next() {
		var sq SimilarQuery
		if err := rows.Scan(
			&sq.ID,
			&sq.QueryText,
			&sq.ExecutionTimeMs,
			&sq.RowsReturned,
			&sq.ExecutedAt,
			&sq.TextScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan FTS result: %w", err)
		}
		sq.VectorScore = 0 // Not from vector search
		results = append(results, &sq)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("FTS row iteration error: %w", err)
	}

	return results, nil
}

// vectorSearch performs semantic similarity search using pgvector.
func (hss *HybridSearchService) vectorSearch(ctx context.Context, embedding []float32, limit int) ([]*SimilarQuery, error) {
	if len(embedding) == 0 {
		return []*SimilarQuery{}, nil
	}

	// Convert embedding to PostgreSQL vector format
	// In actual implementation, this would use pgvector-compatible format
	queryStr := `
		SELECT 
			id,
			query_text,
			execution_time_ms,
			rows_returned,
			executed_at,
			1 - (query_embedding <=> $1::vector) as vector_score
		FROM query_logs
		WHERE query_embedding IS NOT NULL
		ORDER BY query_embedding <=> $1::vector
		LIMIT $2
	`

	// Convert float32 slice to vector representation
	// This is a simplified version - actual pgvector binding may differ
	embeddingStr := vectorToString(embedding)

	rows, err := hss.db.QueryContext(ctx, queryStr, embeddingStr, limit)
	if err != nil {
		// Vector search might not be available if pgvector extension not installed
		// Gracefully degrade
		return []*SimilarQuery{}, nil
	}
	defer rows.Close()

	var results []*SimilarQuery
	for rows.Next() {
		var sq SimilarQuery
		if err := rows.Scan(
			&sq.ID,
			&sq.QueryText,
			&sq.ExecutionTimeMs,
			&sq.RowsReturned,
			&sq.ExecutedAt,
			&sq.VectorScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan vector search result: %w", err)
		}
		sq.TextScore = 0 // Not from text search
		results = append(results, &sq)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("vector search iteration error: %w", err)
	}

	return results, nil
}

// generateEmbedding generates semantic embedding for a query using LLM.
// This is a placeholder - actual implementation depends on LLM capabilities.
func (hss *HybridSearchService) generateEmbedding(ctx context.Context, queryText string) ([]float32, error) {
	if hss.llmClient == nil {
		return []float32{}, nil
	}

	// This is a simplified version - in production, you'd call an embeddings API
	// Placeholder: generate zero vector
	embedding := make([]float32, 1536)
	return embedding, nil
}

// vectorToString converts a float32 slice to PostgreSQL vector string format.
// Format: "[0.1, 0.2, ..., 0.9]"
func vectorToString(embedding []float32) string {
	if len(embedding) == 0 {
		return "[]"
	}

	parts := make([]string, len(embedding))
	for i, v := range embedding {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// sortByScore sorts SimilarQuery slices by combined score (highest first).
func sortByScore(queries []*SimilarQuery) {
	// Bubble sort for simplicity - could use sort.Slice for production
	for i := 0; i < len(queries); i++ {
		for j := i + 1; j < len(queries); j++ {
			if queries[j].CombinedScore > queries[i].CombinedScore {
				queries[i], queries[j] = queries[j], queries[i]
			}
		}
	}
}

// GetSearchStats returns statistics about query logs for monitoring.
type SearchStats struct {
	TotalQueries      int
	QueriesWithEmbedding int
	SlowQueryCount    int
	AverageExecutionMs float64
	LastLogTime       *time.Time
}

// GetSearchStats retrieves search statistics for monitoring.
func (hss *HybridSearchService) GetSearchStats(ctx context.Context) (*SearchStats, error) {
	queryStr := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN query_embedding IS NOT NULL THEN 1 ELSE 0 END), 0) as with_embedding,
			COALESCE(SUM(CASE WHEN is_slow THEN 1 ELSE 0 END), 0) as slow_count,
			COALESCE(AVG(execution_time_ms), 0) as avg_exec_ms,
			MAX(executed_at) as last_log
		FROM query_logs
	`

	var stats SearchStats
	var lastLog *time.Time

	err := hss.db.QueryRowContext(ctx, queryStr).Scan(
		&stats.TotalQueries,
		&stats.QueriesWithEmbedding,
		&stats.SlowQueryCount,
		&stats.AverageExecutionMs,
		&lastLog,
	)

	if err != nil {
		// It's OK if no rows exist (empty table)
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get search stats: %w", err)
		}
	}

	stats.LastLogTime = lastLog
	return &stats, nil
}
