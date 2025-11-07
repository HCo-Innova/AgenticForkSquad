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
)

// OperativoCompatAgent focuses on partitioning and schema redesign prompts (Operativo role).
type OperativoCompatAgent struct {
	Base *BaseAgent
	MCPQ mcpQueryPort
	LLM  llm.LLMClient
}

func (a *OperativoCompatAgent) AnalyzeTask(ctx context.Context, task *entities.Task, forkID string) (AnalysisResult, error) {
	if a == nil || a.MCPQ == nil || a.LLM == nil || task == nil {
		return AnalysisResult{}, errors.New("agent not initialized")
	}
	explainSQL := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", task.TargetQuery)
	res, err := a.MCPQ.ExecuteQuery(ctx, forkID, explainSQL, 60000)
	if err != nil { return AnalysisResult{}, err }
	_ = res
	system := "You are a Senior Data Architect (Operativo role). Respond ONLY with JSON."
	prompt := strings.Join([]string{
		"Analyze schema and partitioning opportunities.",
		"Return fields: insights[], issues[], focus_areas[].",
		"Query:", task.TargetQuery,
	}, "\n")
	obj, err := a.LLM.SendMessageWithJSON(prompt, system)
	if err != nil { return AnalysisResult{}, err }
	ar := AnalysisResult{}
	if v, ok := obj["insights"].([]interface{}); ok { ar.Insights = toStringSlice(v) }
	if v, ok := obj["issues"].([]interface{}); ok { ar.Issues = toStringSlice(v) }
	if v, ok := obj["focus_areas"].([]interface{}); ok { ar.Focus = toStringSlice(v) }
	return ar, nil
}

func (a *OperativoCompatAgent) ProposeOptimization(ctx context.Context, analysis AnalysisResult, forkID string) (*entities.OptimizationProposal, error) {
	if a == nil || a.LLM == nil { return nil, errors.New("agent not initialized") }
	system := "You are an Operativo agent. Propose partitioning or schema redesign. JSON only."
	prompt := "Output fields: proposal_type, sql_commands[], rationale"
	obj, err := a.LLM.SendMessageWithJSON(prompt, system)
	if err != nil { return nil, err }
	typeStr := getString(obj, "proposal_type")
	cmds := getStringSlice(obj, "sql_commands")
	rat := getString(obj, "rationale")
	est := entities.EstimatedImpact{QueryTimeImprovement: 8, StorageOverheadMB: 2, Complexity: "medium", Risk: "medium"}
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

func (a *OperativoCompatAgent) RunBenchmark(ctx context.Context, proposal *entities.OptimizationProposal, forkID string) ([]*entities.BenchmarkResult, error) {
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
			total += qr.ExecutionTimeMs
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
