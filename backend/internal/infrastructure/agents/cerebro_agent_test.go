package agents

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

type mockLLM_G struct { jsonResp map[string]interface{} }
func (m *mockLLM_G) SendMessage(prompt, system string) (string, error) { return "", nil }
func (m *mockLLM_G) SendMessageWithJSON(prompt, system string) (map[string]interface{}, error) {
	b, _ := json.Marshal(m.jsonResp)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out, nil
}
func (m *mockLLM_G) GetUsage() (int, int) { return 0, 0 }

type mockMCPQ_G struct{ calls int; times []float64 }
func (m *mockMCPQ_G) ExecuteQuery(ctx context.Context, serviceID, sql string, timeoutMs int) (mcp.QueryResult, error) {
	val := 10.0
	if m.calls < len(m.times) { val = m.times[m.calls] }
	m.calls++
	return mcp.QueryResult{ExecutionTimeMs: val}, nil
}

func TestAgentCerebro(t *testing.T) {
	mcpq := &mockMCPQ_G{times: []float64{11,10,9, 21,20,19, 31,30,29, 41,40,39}}
	llm := &mockLLM_G{jsonResp: map[string]interface{}{
		"insights": []interface{}{"materialized view candidate"},
		"issues": []interface{}{"expensive aggregate"},
		"focus_areas": []interface{}{"mv"},
	}}
	ag := &CerebroAgent{MCPQ: mcpq, LLM: llm}

	task := &entities.Task{TargetQuery: "SELECT * FROM orders WHERE status='completed'"}
	ar, err := ag.AnalyzeTask(context.Background(), task, "fork-1")
	if err != nil { t.Fatalf("AnalyzeTask err: %v", err) }
	if len(ar.Insights) == 0 { t.Fatalf("expected insights") }

	llm.jsonResp = map[string]interface{}{
		"proposal_type": "materialized_view",
		"sql_commands": []interface{}{"CREATE MATERIALIZED VIEW mv AS SELECT * FROM orders;"},
		"rationale": "precompute",
	}
	prop, err := ag.ProposeOptimization(context.Background(), ar, "fork-1")
	if err != nil { t.Fatalf("ProposeOptimization err: %v", err) }
	prop.ID = 1

	res, err := ag.RunBenchmark(context.Background(), prop, "fork-1")
	if err != nil { t.Fatalf("RunBenchmark err: %v", err) }
	if len(res) != 4 { t.Fatalf("expected 4 results, got %d", len(res)) }
}
