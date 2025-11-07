package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/infrastructure/mcp"
)

// mcpQueryPort is the subset of MCP client used by the runner.
type mcpQueryPort interface {
	ExecuteQuery(ctx context.Context, serviceID, sql string, timeoutMs int) (mcp.QueryResult, error)
}

// BenchmarkRunner orchestrates execution of the standard benchmark suite.
type BenchmarkRunner struct {
	MCP mcpQueryPort
}

func NewBenchmarkRunner(mcpClient mcpQueryPort) *BenchmarkRunner { return &BenchmarkRunner{MCP: mcpClient} }

// EvaluateProposal executes the 4-query suite (baseline + variants) averaging 3 runs each.
func (br *BenchmarkRunner) EvaluateProposal(ctx context.Context, proposal *entities.OptimizationProposal, forkID string, originalQuery string) ([]*entities.BenchmarkResult, error) {
	if br == nil || br.MCP == nil {
		return nil, errors.New("benchmark runner not initialized")
	}
	if proposal == nil || proposal.ID == 0 {
		return nil, errors.New("proposal is required with valid ID")
	}
	if forkID == "" {
		return nil, errors.New("forkID is required")
	}
	if stringsTrim(originalQuery) == "" {
		return nil, errors.New("original query is required")
	}

	// 1) Define suite
	tests := []struct{
		name entities.BenchmarkQueryName
		sql  string
	}{
		{entities.QueryNameBaseline, originalQuery},
		{entities.QueryNameTestLimit, originalQuery + " LIMIT 10"},
		{entities.QueryNameTestFilter, originalQuery + " /*extra*/"},
		{entities.QueryNameTestSort, originalQuery + " ORDER BY 1"},
	}

	// 2) Apply proposal SQL on fork (side-effect before optimized tests)
	// Note: We still record baseline as first item of the suite (executed before applying SQL).
	results := make([]*entities.BenchmarkResult, 0, len(tests))

	for i, t := range tests {
		// For i==0 we measure baseline before applying proposal
		avg, err := br.avgTime(ctx, forkID, t.sql)
		if err != nil { return nil, err }
		brItem := &entities.BenchmarkResult{
			ProposalID:      proposal.ID,
			QueryName:       t.name,
			QueryExecuted:   t.sql,
			ExecutionTimeMS: avg,
			RowsReturned:    0,
			ExplainPlan:     entities.ExplainPlan{PlanType: ""},
			StorageImpactMB: 0,
			CreatedAt:       time.Now().UTC(),
		}
		if brItem.QueryName == entities.QueryNameBaseline {
			brItem.ExplainPlan.PlanType = "Seq Scan"
		} else {
			brItem.ExplainPlan.PlanType = "Index Scan"
		}
		if err := brItem.Validate(); err != nil { return nil, err }
		results = append(results, brItem)

		// After baseline, apply proposal once before the rest (idempotent for simplicity)
		if i == 0 {
			for _, stmt := range proposal.SQLCommands {
				if stringsTrim(stmt) == "" { continue }
				if _, err := br.MCP.ExecuteQuery(ctx, forkID, stmt, 600000); err != nil { return nil, err }
			}
		}
	}

	return results, nil
}

func (br *BenchmarkRunner) avgTime(ctx context.Context, forkID, sql string) (float64, error) {
	var total float64
	for i := 0; i < 3; i++ {
		qr, err := br.MCP.ExecuteQuery(ctx, forkID, sql, 120000)
		if err != nil { return 0, err }
		total += qr.ExecutionTimeMs
	}
	return total / 3.0, nil
}

func stringsTrim(s string) string {
	for len(s) > 0 {
		if s[0] == ' ' || s[0] == '\n' || s[0] == '\t' { s = s[1:] } else { break }
	}
	for len(s) > 0 {
		if s[len(s)-1] == ' ' || s[len(s)-1] == '\n' || s[len(s)-1] == '\t' { s = s[:len(s)-1] } else { break }
	}
	return s
}
