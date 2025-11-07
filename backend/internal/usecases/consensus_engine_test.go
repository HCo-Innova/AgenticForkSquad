package usecases

import (
	"context"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

func TestConsensus(t *testing.T) {
	ce := NewConsensusEngine()
	criteria := entities.ScoringCriteria{PerformanceWeight: 0.5, StorageWeight: 0.2, ComplexityWeight: 0.2, RiskWeight: 0.1}

	// Proposals ordered: cerebro, operativo, bulk
	p1 := &entities.OptimizationProposal{ID: 1, EstimatedImpact: entities.EstimatedImpact{StorageOverheadMB: 10, Risk: "low"}}
	p2 := &entities.OptimizationProposal{ID: 2, EstimatedImpact: entities.EstimatedImpact{StorageOverheadMB: 40, Risk: "medium"}}
	p3 := &entities.OptimizationProposal{ID: 3, EstimatedImpact: entities.EstimatedImpact{StorageOverheadMB: 50, Risk: "medium"}}
	props := []*entities.OptimizationProposal{p1, p2, p3}

	// Benchmarks (baseline + best optimized) to hit target improvements
	bms := []*entities.BenchmarkResult{
		// p1: improvement 90%
		{ProposalID: 1, QueryName: entities.QueryNameBaseline, ExecutionTimeMS: 100.0},
		{ProposalID: 1, QueryName: entities.QueryNameTestLimit, ExecutionTimeMS: 10.0},
		// p2: improvement 79%
		{ProposalID: 2, QueryName: entities.QueryNameBaseline, ExecutionTimeMS: 100.0},
		{ProposalID: 2, QueryName: entities.QueryNameTestLimit, ExecutionTimeMS: 21.0},
		// p3: improvement 59%
		{ProposalID: 3, QueryName: entities.QueryNameBaseline, ExecutionTimeMS: 100.0},
		{ProposalID: 3, QueryName: entities.QueryNameTestLimit, ExecutionTimeMS: 41.0},
	}

	dec, err := ce.Decide(context.Background(), props, bms, criteria)
	if err != nil { t.Fatalf("Decide err: %v", err) }
	if dec == nil { t.Fatalf("nil decision") }

	// Extract scores
	s1 := dec.AllScores[values.AgentCerebro]
	s2 := dec.AllScores[values.AgentOperativo]
	s3 := dec.AllScores[values.AgentBulk]

	// Expected totals from doc example (allow small rounding tolerance)
	expect := func(got, want float64) {
		if got < want-0.1 || got > want+0.1 {
			t.Fatalf("expected %.1f, got %.2f", want, got)
		}
	}
	expect(s1.WeightedTotal, 93.0)
	expect(s2.WeightedTotal, 78.5)
	expect(s3.WeightedTotal, 66.5)

	// Winner should be cerebro (rank 1)
	if s1.Rank != 1 {
		t.Fatalf("expected cerebro rank 1, got %d", s1.Rank)
	}
}
