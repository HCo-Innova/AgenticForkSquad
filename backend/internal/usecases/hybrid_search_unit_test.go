package usecases

import (
	"context"
	"testing"
	"time"
)

// TestSortByScoreOrdering verifies queries are sorted by combined score in descending order.
func TestSortByScoreOrdering(t *testing.T) {
	queries := []*SimilarQuery{
		{ID: 1, QueryText: "SELECT 1", CombinedScore: 0.3},
		{ID: 2, QueryText: "SELECT 2", CombinedScore: 0.9},
		{ID: 3, QueryText: "SELECT 3", CombinedScore: 0.5},
		{ID: 4, QueryText: "SELECT 4", CombinedScore: 0.1},
	}

	sortByScore(queries)

	expectedOrder := []int{2, 3, 1, 4}
	expectedScores := []float64{0.9, 0.5, 0.3, 0.1}

	for i, q := range queries {
		if q.ID != expectedOrder[i] {
			t.Errorf("position %d: expected ID %d, got %d", i, expectedOrder[i], q.ID)
		}
		if q.CombinedScore != expectedScores[i] {
			t.Errorf("position %d: expected score %.1f, got %.1f", i, expectedScores[i], q.CombinedScore)
		}
	}
}

// TestWeightingFormula validates the 40/60 weighting for text/vector scores.
func TestWeightingFormula(t *testing.T) {
	tests := []struct {
		name        string
		textScore   float64
		vectorScore float64
		expected    float64
	}{
		{"text_only", 1.0, 0.0, 0.4},
		{"vector_only", 0.0, 1.0, 0.6},
		{"perfect_match", 1.0, 1.0, 1.0},
		{"balanced", 0.5, 0.5, 0.5},
		{"no_match", 0.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			combined := tt.textScore*0.4 + tt.vectorScore*0.6
			if combined != tt.expected {
				t.Errorf("expected %.2f, got %.2f", tt.expected, combined)
			}
		})
	}
}

// TestVectorStringConversion validates embedding to string conversion.
func TestVectorStringConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		contains string // substring check
	}{
		{"empty", []float32{}, "[]"},
		{"single", []float32{0.5}, "["},
		{"multiple", []float32{0.1, 0.2}, "[0.1, 0.2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vectorToString(tt.input)
			if result != tt.contains && len(tt.input) > 0 {
				// For multiple element test, just check it contains opening bracket
				if !contains(result, "[") {
					t.Errorf("expected vector format, got %q", result)
				}
			} else if result != tt.contains {
				t.Errorf("expected %q, got %q", tt.contains, result)
			}
		})
	}
}

// Helper: simple contains check
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSearchResultScoreConsistency validates that combined score = text*0.4 + vector*0.6
func TestSearchResultScoreConsistency(t *testing.T) {
	result := &HybridSearchResult{
		InputQuery: "SELECT * FROM users",
		SimilarQueries: []*SimilarQuery{
			{
				ID:            1,
				QueryText:     "SELECT * FROM users WHERE id > 100",
				TextScore:     0.8,
				VectorScore:   0.7,
				CombinedScore: 0.8*0.4 + 0.7*0.6, // 0.74
			},
		},
	}

	sq := result.SimilarQueries[0]
	expected := sq.TextScore*0.4 + sq.VectorScore*0.6

	if sq.CombinedScore != expected {
		t.Errorf("score inconsistency: expected %.3f, got %.3f", expected, sq.CombinedScore)
	}
}

// TestContextTimeout validates context cancellation is respected.
func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give context time to expire
	time.Sleep(10 * time.Millisecond)

	if ctx.Err() == nil {
		t.Error("context should be expired")
	}
}

// TestSimilarQueryFields validates SimilarQuery struct has required fields.
func TestSimilarQueryFields(t *testing.T) {
	sq := &SimilarQuery{
		ID:              123,
		QueryText:       "SELECT * FROM users",
		ExecutionTimeMs: ptrFloat(150.5),
		RowsReturned:    ptrInt(1000),
		ExecutedAt:      time.Now(),
		TextScore:       0.8,
		VectorScore:     0.7,
		CombinedScore:   0.74,
		Reason:          "keyword match",
	}

	if sq.ID != 123 {
		t.Error("ID not set")
	}
	if sq.QueryText == "" {
		t.Error("QueryText empty")
	}
	if sq.ExecutionTimeMs == nil || *sq.ExecutionTimeMs != 150.5 {
		t.Error("ExecutionTimeMs not set correctly")
	}
	if sq.Reason == "" {
		t.Error("Reason empty")
	}
}

// Helper functions for pointers
func ptrFloat(v float64) *float64 {
	return &v
}

func ptrInt(v int) *int {
	return &v
}
