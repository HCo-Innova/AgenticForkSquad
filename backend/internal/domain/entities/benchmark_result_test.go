package entities

import (
	"testing"
	"time"
)

func TestValidate_ValidBenchmark(t *testing.T) {
	b := &BenchmarkResult{
		ProposalID:      1,
		QueryName:       QueryNameBaseline,
		QueryExecuted:   "SELECT * FROM users WHERE id = 1;",
		ExecutionTimeMS: 25.3,
		RowsReturned:    1,
		StorageImpactMB: 10.5,
		ExplainPlan: ExplainPlan{
			PlanType: "Index Scan",
			TotalCost: 1.2,
			ActualRows: 1,
		},
		CreatedAt: time.Now(),
	}

	if err := b.Validate(); err != nil {
		t.Fatalf("expected valid benchmark, got error: %v", err)
	}
}

func TestValidate_MissingQueryName(t *testing.T) {
	b := &BenchmarkResult{
		ProposalID:      1,
		QueryName:       "",
		QueryExecuted:   "SELECT * FROM users;",
		ExecutionTimeMS: 10,
		ExplainPlan:     ExplainPlan{PlanType: "test"},
	}
	if err := b.Validate(); err == nil {
		t.Fatal("expected error for missing query name")
	}
}

func TestValidate_NegativeExecutionTime(t *testing.T) {
	b := &BenchmarkResult{
		ProposalID:      1,
		QueryName:       QueryNameTestFilter,
		QueryExecuted:   "SELECT * FROM x;",
		ExecutionTimeMS: -1,
		ExplainPlan:     ExplainPlan{PlanType: "plan"},
	}
	if err := b.Validate(); err == nil {
		t.Fatal("expected error for negative execution time")
	}
}

func TestValidate_EmptyExplainPlan(t *testing.T) {
	b := &BenchmarkResult{
		ProposalID:      1,
		QueryName:       QueryNameTestSort,
		QueryExecuted:   "SELECT 1;",
		ExecutionTimeMS: 10,
	}
	if err := b.Validate(); err == nil {
		t.Fatal("expected error for missing explain plan details")
	}
}

func TestIsFasterThan(t *testing.T) {
	a := &BenchmarkResult{ExecutionTimeMS: 10}
	b := &BenchmarkResult{ExecutionTimeMS: 20}

	if !a.IsFasterThan(b) {
		t.Error("expected benchmark A to be faster than B")
	}

	if b.IsFasterThan(a) {
		t.Error("expected benchmark B not to be faster than A")
	}

	if a.IsFasterThan(nil) {
		t.Error("expected false when comparing with nil")
	}
}

func TestEfficiencyRatio(t *testing.T) {
	b := &BenchmarkResult{ExecutionTimeMS: 20, StorageImpactMB: 5}
	if b.EfficiencyRatio() != 4 {
		t.Errorf("expected 4, got %.2f", b.EfficiencyRatio())
	}

	b.StorageImpactMB = 0
	if b.EfficiencyRatio() != 0 {
		t.Error("expected 0 when storage impact is 0")
	}
}