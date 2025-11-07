package usecases

import (
	"context"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

type mockMCPRunner struct{ calls int; times []float64 }
func (m *mockMCPRunner) ExecuteQuery(ctx context.Context, serviceID, sql string, timeoutMs int) (mcp.QueryResult, error) {
	val := 10.0
	if m.calls < len(m.times) { val = m.times[m.calls] }
	m.calls++
	return mcp.QueryResult{ExecutionTimeMs: val}, nil
}

func TestBenchmarkRunner(t *testing.T) {
	// 12 timings for 4 queries × 3 runs each; extra calls will be for applying SQL
	m := &mockMCPRunner{times: []float64{12, 9, 15, 20, 22, 21, 30, 31, 29, 40, 39, 41}}
	r := NewBenchmarkRunner(m)
	prop := &entities.OptimizationProposal{ID: 1, SQLCommands: []string{"CREATE INDEX x ON orders(status)", "ANALYZE orders"}}
	orig := "SELECT * FROM orders"
	res, err := r.EvaluateProposal(context.Background(), prop, "fork-1", orig)
	if err != nil { t.Fatalf("EvaluateProposal err: %v", err) }
	if len(res) != 4 { t.Fatalf("expected 4 results, got %d", len(res)) }
	// Validate first avg = (12+9+15)/3 = 12
	if res[0].ExecutionTimeMS < 11.9 || res[0].ExecutionTimeMS > 12.1 {
		t.Fatalf("unexpected avg for baseline: %v", res[0].ExecutionTimeMS)
	}
	// Verify total query calls: 12 (measurements) + 2 (apply SQL)
	if m.calls < 14 { t.Fatalf("expected at least 14 calls (12 runs + 2 apply), got %d", m.calls) }
}
