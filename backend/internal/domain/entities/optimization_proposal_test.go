package entities

import (
	"testing"
	"time"
)

func TestValidate_ValidProposal(t *testing.T) {
	p := &OptimizationProposal{
		ID:               1,
		AgentExecutionID: 10,
		ProposalType:     ProposalTypeIndex,
		SQLCommands:      []string{"CREATE INDEX idx_users_email ON users(email);"},
		Rationale:        "Improve lookup performance by email.",
		EstimatedImpact: EstimatedImpact{
			QueryTimeImprovement: 35.4,
			StorageOverheadMB:    12.7,
			Complexity:           "low",
			Risk:                 "low",
		},
		CreatedAt: time.Now(),
	}

	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid proposal, got: %v", err)
	}
}

func TestValidate_InvalidProposalType(t *testing.T) {
	p := &OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     "invalid_type",
		SQLCommands:      []string{"SELECT 1;"},
		Rationale:        "test",
		EstimatedImpact:  EstimatedImpact{QueryTimeImprovement: 5, Risk: "low", Complexity: "low"},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for invalid proposal type")
	}
}

func TestValidate_EmptySQLCommands(t *testing.T) {
	p := &OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     ProposalTypePartitioning,
		SQLCommands:      []string{},
		Rationale:        "test",
		EstimatedImpact:  EstimatedImpact{QueryTimeImprovement: 10, Risk: "medium", Complexity: "high"},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty SQL commands")
	}
}

func TestValidate_EmptyStatementInsideSQLCommands(t *testing.T) {
	p := &OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     ProposalTypeIndex,
		SQLCommands:      []string{"CREATE INDEX idx;", " "},
		Rationale:        "test",
		EstimatedImpact:  EstimatedImpact{QueryTimeImprovement: 5, Risk: "low", Complexity: "low"},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty SQL command entry")
	}
}

func TestValidate_MissingRationale(t *testing.T) {
	p := &OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     ProposalTypeMaterializedView,
		SQLCommands:      []string{"CREATE MATERIALIZED VIEW mv_users AS SELECT * FROM users;"},
		Rationale:        "",
		EstimatedImpact:  EstimatedImpact{QueryTimeImprovement: 5, Risk: "low", Complexity: "medium"},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing rationale")
	}
}

func TestValidate_InvalidRiskLevel(t *testing.T) {
	p := &OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     ProposalTypeQueryRewrite,
		SQLCommands:      []string{"EXPLAIN ANALYZE SELECT * FROM users;"},
		Rationale:        "Optimize complex joins",
		EstimatedImpact:  EstimatedImpact{QueryTimeImprovement: 5, Risk: "extreme", Complexity: "low"},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for invalid risk level")
	}
}

func TestValidate_ZeroPerformanceGain(t *testing.T) {
	p := &OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     ProposalTypeDenormalization, // Using one of the new types
		SQLCommands:      []string{"ALTER TABLE users ADD COLUMN age INT;"},
		Rationale:        "Add age field for analytics",
		EstimatedImpact:  EstimatedImpact{QueryTimeImprovement: 0, Risk: "medium", Complexity: "medium"},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for zero performance gain")
	}
}

func TestValidate_MissingComplexity(t *testing.T) {
	p := &OptimizationProposal{
		AgentExecutionID: 1,
		ProposalType:     ProposalTypeIndex,
		SQLCommands:      []string{"CREATE INDEX idx_test ON users(email);"},
		Rationale:        "A valid rationale.",
		EstimatedImpact: EstimatedImpact{
			QueryTimeImprovement: 20,
			Risk:                 "low",
			// Complexity is missing
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing complexity level")
	}
}