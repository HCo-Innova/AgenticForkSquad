package agents

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

// ---- Mocks ----

type mockLLM struct {
	jsonResp map[string]interface{}
	text    string
}

func (m *mockLLM) SendMessage(prompt, system string) (string, error) {
	return m.text, nil
}
func (m *mockLLM) SendMessageWithJSON(prompt, system string) (map[string]interface{}, error) {
	// Return a deep copy to avoid mutation between calls
	b, _ := json.Marshal(m.jsonResp)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out, nil
}
func (m *mockLLM) GetUsage() (int, int) { return 0, 0 }

type mockMCPQuery struct{ calls int; times []float64 }
func (m *mockMCPQuery) ExecuteQuery(ctx context.Context, serviceID, sql string, timeoutMs int) (mcp.QueryResult, error) {
	val := 10.0
	if m.calls < len(m.times) { val = m.times[m.calls] }
	m.calls++
	return mcp.QueryResult{ExecutionTimeMs: val}, nil
}

// ---- Tests ----

func TestAgentOperativo(t *testing.T) {
	// Prepare mocks
	mcpq := &mockMCPQuery{times: []float64{12, 9, 15, 20, 22, 21, 30, 28, 29, 40, 41, 39}}
	llm := &mockLLM{jsonResp: map[string]interface{}{
		"insights": []interface{}{"use index"},
		"issues": []interface{}{"seq scan"},
		"focus_areas": []interface{}{"filter"},
	}}
	ag := &OperativoAgent{MCPQ: mcpq, LLM: llm}

	// AnalyzeTask
	task := &entities.Task{TargetQuery: "SELECT * FROM orders WHERE status='completed'"}
	ar, err := ag.AnalyzeTask(context.Background(), task, "fork-1")
	if err != nil { t.Fatalf("AnalyzeTask err: %v", err) }
	if len(ar.Insights) == 0 || len(ar.Issues) == 0 { t.Fatalf("analysis empty: %+v", ar) }

	// ProposeOptimization
	llm.jsonResp = map[string]interface{}{
		"proposal_type": "index",
		"sql_commands": []interface{}{"CREATE INDEX idx_orders_status ON orders(status)"},
		"rationale": "common filter",
	}
	prop, err := ag.ProposeOptimization(context.Background(), ar, "fork-1")
	if err != nil { t.Fatalf("ProposeOptimization err: %v", err) }
	if len(prop.SQLCommands) == 0 { t.Fatalf("expected sql commands") }
	// Ensure ProposalID is positive for benchmark validation
	prop.ID = 1

	// RunBenchmark: should return 4 results with averaged times (based on mock sequence)
	res, err := ag.RunBenchmark(context.Background(), prop, "fork-1")
	if err != nil { t.Fatalf("RunBenchmark err: %v", err) }
	if len(res) != 4 { t.Fatalf("expected 4 results, got %d", len(res)) }
	// First avg should be (12+9+15)/3 = 12
	if res[0].ExecutionTimeMS <= 0 { t.Fatalf("unexpected avg: %v", res[0].ExecutionTimeMS) }
}
