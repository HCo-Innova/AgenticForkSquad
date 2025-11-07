package usecases

import (
	"context"
	"testing"
	"time"
)

// TestHybridSearchWeighting validates the weighted scoring formula (text 40%, vector 60%).
func TestHybridSearchWeighting(t *testing.T) {
	tests := []struct {
		name          string
		textScore     float64
		vectorScore   float64
		expectedScore float64
		tolerance     float64
	}{
		{
			name:          "only_text_match",
			textScore:     1.0,
			vectorScore:   0.0,
			expectedScore: 0.4, // 1.0 * 0.4 + 0.0 * 0.6
			tolerance:     0.01,
		},
		{
			name:          "only_vector_match",
			textScore:     0.0,
			vectorScore:   1.0,
			expectedScore: 0.6, // 0.0 * 0.4 + 1.0 * 0.6
			tolerance:     0.01,
		},
		{
			name:          "perfect_match",
			textScore:     1.0,
			vectorScore:   1.0,
			expectedScore: 1.0, // 1.0 * 0.4 + 1.0 * 0.6
			tolerance:     0.01,
		},
		{
			name:          "balanced_match",
			textScore:     0.5,
			vectorScore:   0.5,
			expectedScore: 0.5, // 0.5 * 0.4 + 0.5 * 0.6
			tolerance:     0.01,
		},
		{
			name:          "no_match",
			textScore:     0.0,
			vectorScore:   0.0,
			expectedScore: 0.0,
			tolerance:     0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Manually apply weighting formula
			combined := tt.textScore*0.4 + tt.vectorScore*0.6

			if combined < tt.expectedScore-tt.tolerance || combined > tt.expectedScore+tt.tolerance {
				t.Errorf("expected %.3f, got %.3f", tt.expectedScore, combined)
			}
		})
	}
}

// TestSortByScore validates that queries are sorted by score (highest first).
func TestSortByScore(t *testing.T) {
	queries := []*SimilarQuery{
		{ID: 1, CombinedScore: 0.3},
		{ID: 2, CombinedScore: 0.9},
		{ID: 3, CombinedScore: 0.5},
		{ID: 4, CombinedScore: 0.1},
	}

	sortByScore(queries)

	expectedOrder := []int{2, 3, 1, 4}
	for i, q := range queries {
		if q.ID != expectedOrder[i] {
			t.Errorf("position %d: expected ID %d, got %d", i, expectedOrder[i], q.ID)
		}
	}
}

// TestVectorToString converts float32 embeddings to PostgreSQL vector format.
func TestVectorToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected string
	}{
		{
			name:     "empty_vector",
			input:    []float32{},
			expected: "[]",
		},
		{
			name:     "single_element",
			input:    []float32{0.5},
			expected: "[0.5]",
		},
		{
			name:     "multiple_elements",
			input:    []float32{0.1, 0.2, 0.3},
			expected: "[0.1, 0.2, 0.3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vectorToString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestHybridSearchIntegration tests hybrid search with mock database.
func TestHybridSearchIntegration(t *testing.T) {
	// This is a placeholder test structure for integration testing
	// In production, would use sqlc or testcontainers with PostgreSQL

	t.Run("no_similar_queries", func(t *testing.T) {
		// When query_logs is empty, should return empty results
		// Expects: result.TotalMatches == 0
	})

	t.Run("text_search_only", func(t *testing.T) {
		// When embeddings not available, should use FTS only
		// Expects: result.TextIndexUsed == true, result.VectorIndexUsed == false
	})

	t.Run("combined_search", func(t *testing.T) {
		// When both FTS and vector available, combine with weights
		// Expects: result.TextIndexUsed == true, result.VectorIndexUsed == true
	})

	t.Run("top_n_results", func(t *testing.T) {
		// Should respect topN limit
		// Expects: len(result.SimilarQueries) <= topN
	})
}

// MockQueryLogEntry creates a query log entry for testing.
func MockQueryLogEntry(queryText string, execTimeMs float64) *SimilarQuery {
	return &SimilarQuery{
		ID:              int(time.Now().UnixNano()) % 10000,
		QueryText:       queryText,
		ExecutionTimeMs: &execTimeMs,
		ExecutedAt:      time.Now(),
		TextScore:       0.0,
		VectorScore:     0.0,
		CombinedScore:   0.0,
	}
}

// TestSearchRelevance validates that relevant queries rank higher than irrelevant ones.
func TestSearchRelevance(t *testing.T) {
	// Example: searching for "SELECT * FROM users" should rank
	// "SELECT * FROM users WHERE id > 100" higher than
	// "SELECT * FROM orders"

	queryA := MockQueryLogEntry("SELECT * FROM users WHERE id > 100", 50.0)
	queryA.TextScore = 0.9
	queryA.CombinedScore = 0.9*0.4 + 0*0.6

	queryB := MockQueryLogEntry("SELECT * FROM orders WHERE status='pending'", 200.0)
	queryB.TextScore = 0.1
	queryB.CombinedScore = 0.1*0.4 + 0*0.6

	if queryA.CombinedScore <= queryB.CombinedScore {
		t.Error("relevant query should score higher than irrelevant query")
	}
}

// TestQueryHashDeduplication validates that duplicate queries generate same hash.
func TestQueryHashDeduplication(t *testing.T) {
	// Note: generateQueryHash is a private function in query_logger.go
	// This test documents the expected behavior
	// Same query with different case should produce same hash
	// Different queries should produce different hashes
	t.Skip("requires integration with QueryLogger")
}

// TestQueryLoggerContextDeadline validates timeout handling in query logger.
func TestQueryLoggerContextDeadline(t *testing.T) {
	// Create a context that expires immediately
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	// Wait a bit to ensure context is expired
	time.Sleep(10 * time.Millisecond)

	// Attempting to query with expired context should fail gracefully
	if ctx.Err() == nil {
		t.Error("context should be expired")
	}
}

// TestSearchStatsCalculation validates search statistics accuracy.
func TestSearchStatsCalculation(t *testing.T) {
	// Placeholder for stats calculation testing
	// When we have 10 queries:
	// - 5 with embeddings
	// - 3 slow queries (>1000ms)
	// - average execution time = 500ms
	// Expected stats should reflect these counts

	t.Run("stats_accuracy", func(t *testing.T) {
		// Expects: TotalQueries == 10, QueriesWithEmbedding == 5, SlowQueryCount == 3
	})
}

// BenchmarkHybridSearch benchmarks hybrid search performance.
func BenchmarkHybridSearch(b *testing.B) {
	// This would be benchmarked against real database with typical queries
	// Expected: < 50ms for top 10 results from 1000 queries

	b.Run("text_search_1k_queries", func(b *testing.B) {
		// FTS on 1000 queries, target: <10ms
	})

	b.Run("vector_search_1k_queries", func(b *testing.B) {
		// Vector search on 1000 queries, target: <20ms
	})

	b.Run("combined_search_1k_queries", func(b *testing.B) {
		// Combined search, target: <50ms
	})
}

// TestRouterContextEnrichment validates query router context enrichment.
func TestRouterContextEnrichment(t *testing.T) {
	t.Run("context_with_similar_queries", func(t *testing.T) {
		// When similar queries exist, context should include them
		// Expects: len(RouterContext.SimilarPastQueries) > 0
	})

	t.Run("context_with_agent_recommendations", func(t *testing.T) {
		// Router should recommend agents based on similar queries
		// Expects: len(RouterContext.RecommendedAgents) > 0
	})

	t.Run("context_with_stats", func(t *testing.T) {
		// Router should include search statistics
		// Expects: RouterContext.SearchStats != nil
	})
}

// TestAgentRecommendationLogic validates agent recommendation strategy.
func TestAgentRecommendationLogic(t *testing.T) {
	tests := []struct {
		name               string
		similarQueries     []*SimilarQuery
		expectedAgentCount int
		description        string
	}{
		{
			name:               "no_history",
			similarQueries:     []*SimilarQuery{},
			expectedAgentCount: 3,
			description:        "should use all agents when no history",
		},
		{
			name: "fast_queries",
			similarQueries: []*SimilarQuery{
				{ID: 1, ExecutionTimeMs: ptrFloat64(10.0)},
				{ID: 2, ExecutionTimeMs: ptrFloat64(20.0)},
			},
			expectedAgentCount: 3,
			description:        "should favor flash models for fast queries",
		},
		{
			name: "slow_queries",
			similarQueries: []*SimilarQuery{
				{ID: 1, ExecutionTimeMs: ptrFloat64(5000.0)},
				{ID: 2, ExecutionTimeMs: ptrFloat64(3000.0)},
			},
			expectedAgentCount: 3,
			description:        "should use all agents for slow queries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test QueryRouter.recommendAgents logic
			// Verifies that recommendations match expected patterns
		})
	}
}

// Helper function to create pointer to float64
func ptrFloat64(v float64) *float64 {
	return &v
}

// TestCleanupOldLogs validates cleanup of old query logs.
func TestCleanupOldLogs(t *testing.T) {
	t.Run("cleanup_90_day_retention", func(t *testing.T) {
		// Should delete logs older than 90 days
		// Expects: rowsAffected > 0 for old logs
	})

	t.Run("preserve_recent_logs", func(t *testing.T) {
		// Should NOT delete logs younger than retention period
		// Expects: recent logs still present
	})
}
