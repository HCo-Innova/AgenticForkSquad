package agents

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

type mockLLM_M struct { jsonResp map[string]interface{} }
func (m *mockLLM_M) SendMessage(prompt, system string) (string, error) { return "", nil }
func (m *mockLLM_M) SendMessageWithJSON(prompt, system string) (map[string]interface{}, error) {
	b, _ := json.Marshal(m.jsonResp)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out, nil
}
func (m *mockLLM_M) GetUsage() (int, int) { return 0, 0 }

type mockMCPQ_M struct{ calls int; times []float64 }
func (m *mockMCPQ_M) ExecuteQuery(ctx context.Context, serviceID, sql string, timeoutMs int) (mcp.QueryResult, error) {
	val := 10.0
	if m.calls < len(m.times) { val = m.times[m.calls] }
	m.calls++
	return mcp.QueryResult{ExecutionTimeMs: val}, nil
}

func TestAgentOperativoCompat(t *testing.T) {
	mcpq := &mockMCPQ_M{times: []float64{10,11,9, 20,21,19, 30,31,29, 40,41,39}}
	llm := &mockLLM_M{jsonResp: map[string]interface{}{
		"insights": []interface{}{"partition by status"},
		"issues": []interface{}{"skew"},
		"focus_areas": []interface{}{"schema"},
	}}
	ag := &OperativoCompatAgent{MCPQ: mcpq, LLM: llm}

	task := &entities.Task{TargetQuery: "SELECT * FROM orders WHERE status='completed'"}
	ar, err := ag.AnalyzeTask(context.Background(), task, "fork-1")
	if err != nil { t.Fatalf("AnalyzeTask err: %v", err) }
	if len(ar.Insights) == 0 { t.Fatalf("expected insights") }

	llm.jsonResp = map[string]interface{}{
		"proposal_type": "partitioning",
		"sql_commands": []interface{}{"ALTER TABLE orders PARTITION BY LIST (status);"},
		"rationale": "hot partition",
	}
	prop, err := ag.ProposeOptimization(context.Background(), ar, "fork-1")
	if err != nil { t.Fatalf("ProposeOptimization err: %v", err) }
	prop.ID = 1

	res, err := ag.RunBenchmark(context.Background(), prop, "fork-1")
	if err != nil { t.Fatalf("RunBenchmark err: %v", err) }
	if len(res) != 4 { t.Fatalf("expected 4 results, got %d", len(res)) }
}
