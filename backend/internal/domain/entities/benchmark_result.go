package entities

import (
	"errors"
	"strings"
	"time"
)

// BenchmarkQueryName defines the standard names for queries in a benchmark suite.
type BenchmarkQueryName string

const (
	QueryNameBaseline  BenchmarkQueryName = "baseline"
	QueryNameTestLimit   BenchmarkQueryName = "test_limit"
	QueryNameTestFilter  BenchmarkQueryName = "test_filter"
	QueryNameTestSort    BenchmarkQueryName = "test_sort"
)

// Buffers represents buffer usage from a query plan, mirroring part of the EXPLAIN JSON output.
type Buffers struct {
	SharedHit  int64 `json:"shared_hit"`
	SharedRead int64 `json:"shared_read"`
}

// ExplainPlan represents the structured query execution plan details.
// This struct mirrors the JSONB field used in the database, parsed from EXPLAIN (FORMAT JSON).
type ExplainPlan struct {
	PlanningTimeMS    float64 `json:"planning_time_ms"`
	ExecutionTimeMS   float64 `json:"execution_time_ms"`
	TotalCost         float64 `json:"total_cost"`
	ActualRows        int64   `json:"actual_rows"`
	PlanType          string  `json:"plan_type"`
	IndexName         string  `json:"index_name,omitempty"`
	FilterRemovedRows int64   `json:"filter_removed_rows"`
	SortMethod        string  `json:"sort_method,omitempty"`
	Buffers           Buffers `json:"buffers"`
	FullPlan          string  `json:"full_plan"` // Storing the full JSON plan as a string
}

// BenchmarkResult represents a single performance benchmark of an optimization proposal.
type BenchmarkResult struct {
	ID              int64
	ProposalID      int64
	QueryName       BenchmarkQueryName
	QueryExecuted   string
	ExecutionTimeMS float64 // measured execution time in milliseconds
	RowsReturned    int64
	ExplainPlan     ExplainPlan
	StorageImpactMB float64 // in MB
	CreatedAt       time.Time
}

// Validate enforces domain-level constraints on BenchmarkResult before persistence or processing.
func (b *BenchmarkResult) Validate() error {
	if b.ProposalID <= 0 {
		return errors.New("proposal_id must be positive")
	}

	switch b.QueryName {
	case QueryNameBaseline, QueryNameTestLimit, QueryNameTestFilter, QueryNameTestSort:
		// valid query name
	default:
		return errors.New("invalid query_name")
	}

	if strings.TrimSpace(b.QueryExecuted) == "" {
		return errors.New("query_executed cannot be empty")
	}

	if b.ExecutionTimeMS <= 0 {
		return errors.New("execution_time_ms must be positive")
	}

	if b.RowsReturned < 0 {
		return errors.New("rows_returned cannot be negative")
	}

	// StorageImpactMB can be 0, but not negative.
	if b.StorageImpactMB < 0 {
		return errors.New("storage_impact_mb cannot be negative")
	}

	if b.ExplainPlan.PlanType == "" {
		return errors.New("explain_plan must have a plan_type")
	}

	return nil
}

// IsFasterThan compares two benchmarks by execution time.
func (b *BenchmarkResult) IsFasterThan(other *BenchmarkResult) bool {
	if other == nil {
		return false
	}
	return b.ExecutionTimeMS < other.ExecutionTimeMS
}

// EfficiencyRatio returns a metric of performance gain per MB of storage impact.
// Returns 0 if StorageImpactMB is zero to prevent division by zero.
func (b *BenchmarkResult) EfficiencyRatio() float64 {
	if b.StorageImpactMB == 0 {
		return 0
	}
	return b.ExecutionTimeMS / b.StorageImpactMB
}
