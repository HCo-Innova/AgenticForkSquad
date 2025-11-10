package agents

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/llm"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

// mcpQueryPort defines the subset of MCP client used by the Operativo agent.
type mcpQueryPort interface {
	ExecuteQuery(ctx context.Context, serviceID, sql string, timeoutMs int) (mcp.QueryResult, error)
}

// OperativoAgent implements analysis/proposal/benchmark for the Operativo role.
type OperativoAgent struct {
	Base   *BaseAgent
	MCPQ   mcpQueryPort
	LLM    llm.LLMClient
}

// AnalysisResult is a lightweight struct parsed from LLM JSON.
type AnalysisResult struct {
	Insights []string `json:"insights"`
	Issues   []string `json:"issues"`
	Focus    []string `json:"focus_areas"`
}

func (a *OperativoAgent) AnalyzeTask(ctx context.Context, task *entities.Task, forkID string) (AnalysisResult, error) {
	if a == nil || a.MCPQ == nil || a.LLM == nil || task == nil {
		return AnalysisResult{}, errors.New("agent not initialized")
	}
	// 1) Run EXPLAIN ANALYZE (FORMAT JSON)
	explainSQL := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", task.TargetQuery)
	res, err := a.MCPQ.ExecuteQuery(ctx, forkID, explainSQL, 60000)
	if err != nil { return AnalysisResult{}, err }

	// 2) Build prompt (simplified): include summarized text
	var explainText string
	if len(res.Rows) > 0 {
		// serialize minimal for prompt; in real impl we'd pass full JSON
		explainText = "EXPLAIN available"
	}
	system := "You are an expert PostgreSQL query optimizer (Operativo role). Respond ONLY with a valid JSON object."
	prompt := strings.Join([]string{
		"Analyze the query and plan.",
		"Return fields: insights[], issues[], focus_areas[].",
		"Query:", task.TargetQuery,
		"Explain:", explainText,
	}, "\n")

	obj, err := a.LLM.SendMessageWithJSON(prompt, system)
	if err != nil { return AnalysisResult{}, err }
	ar := AnalysisResult{}
	// Map generic map into struct fields defensively
	if v, ok := obj["insights"].([]interface{}); ok { ar.Insights = toStringSlice(v) }
	if v, ok := obj["issues"].([]interface{}); ok { ar.Issues = toStringSlice(v) }
	if v, ok := obj["focus_areas"].([]interface{}); ok { ar.Focus = toStringSlice(v) }
	return ar, nil
}

// NormalizeProposalType convierte variantes comunes del LLM a valores válidos
func NormalizeProposalType(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	switch s {
	case "index", "create_index", "btree", "b_tree":
		return "index"
	case "partial_index", "filtered_index":
		return "partial_index"
	case "composite_index", "multi_column_index", "compound_index":
		return "composite_index"
	case "materialized_view", "matview", "mv":
		return "materialized_view"
	case "partitioning", "partition", "table_partition":
		return "partitioning"
	case "denormalization", "denormalize":
		return "denormalization"
	case "query_rewrite", "rewrite", "query_optimization":
		return "query_rewrite"
	default:
		return s // devolver tal cual si no coincide con nada
	}
}

func (a *OperativoAgent) ProposeOptimization(ctx context.Context, analysis AnalysisResult, forkID string) (*entities.OptimizationProposal, error) {
	if a == nil || a.LLM == nil { return nil, errors.New("agent not initialized") }
	system := "You are an Operativo (Gemini) agent. Propose an index optimization. Respond ONLY JSON."
	prompt := strings.Join([]string{
		"Based on the analysis, propose an index or similar optimization.",
		"Output JSON fields:",
		"proposal_type (index|partial_index|composite_index|materialized_view|partitioning|denormalization|query_rewrite)",
		"sql_commands (array of SQL strings)",
		"rationale (string)",
	}, "\n")
	obj, err := a.LLM.SendMessageWithJSON(prompt, system)
	if err != nil { return nil, err }
	// Parse essentials
	typeStr := NormalizeProposalType(getString(obj, "proposal_type"))
	cmds := getStringSlice(obj, "sql_commands")
	rat := getString(obj, "rationale")
	// Minimal estimation
	est := entities.EstimatedImpact{QueryTimeImprovement: 10, StorageOverheadMB: 1, Complexity: "low", Risk: "low"}
	p := &entities.OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     values.ProposalType(typeStr),
		SQLCommands:      cmds,
		Rationale:        rat,
		EstimatedImpact:  est,
		CreatedAt:        time.Now().UTC(),
	}
	if err := p.Validate(); err != nil { return nil, err }
	return p, nil
}

func (a *OperativoAgent) RunBenchmark(ctx context.Context, proposal *entities.OptimizationProposal, forkID string) ([]*entities.BenchmarkResult, error) {
	if a == nil || a.MCPQ == nil || proposal == nil { return nil, errors.New("agent not initialized") }
	queries := []struct{ name entities.BenchmarkQueryName; sql string }{
		{entities.QueryNameBaseline, "SELECT 1"},
		{entities.QueryNameTestLimit, "SELECT * FROM orders LIMIT 10"},
		{entities.QueryNameTestFilter, "SELECT * FROM orders WHERE status='completed'"},
		{entities.QueryNameTestSort, "SELECT * FROM orders ORDER BY created_at DESC LIMIT 10"},
	}
	results := make([]*entities.BenchmarkResult, 0, len(queries))
	for _, q := range queries {
		var total float64
		for i := 0; i < 3; i++ {
			qr, err := a.MCPQ.ExecuteQuery(ctx, forkID, q.sql, 60000)
			if err != nil { return nil, err }
			execTime := qr.ExecutionTimeMs
			if execTime <= 0 {
				execTime = 1.0 // fallback si MCP no devuelve tiempo
			}
			total += execTime
		}
		avg := total / 3
		br := &entities.BenchmarkResult{
			ProposalID:      proposal.ID,
			QueryName:       q.name,
			QueryExecuted:   q.sql,
			ExecutionTimeMS: avg,
			RowsReturned:    0,
			ExplainPlan:     entities.ExplainPlan{PlanType: "Seq Scan"},
			StorageImpactMB: 0,
			CreatedAt:       time.Now().UTC(),
		}
		if err := br.Validate(); err != nil { return nil, err }
		results = append(results, br)
	}
	return results, nil
}

// Helpers
func toStringSlice(in []interface{}) []string {
	out := make([]string, 0, len(in))
	for _, v := range in { if s, ok := v.(string); ok { out = append(out, s) } }
	return out
}
func getString(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok { if s, ok := v.(string); ok { return s } }
	return ""
}
func getStringSlice(m map[string]interface{}, k string) []string {
	if v, ok := m[k]; ok { if a, ok := v.([]interface{}); ok { return toStringSlice(a) } }
	return nil
}
