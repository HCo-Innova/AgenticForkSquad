package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

// EstimatedImpact represents expected performance and storage effects
// from applying an optimization. This mirrors the JSONB structure stored in DB.
type EstimatedImpact struct {
	QueryTimeImprovement float64 // Expected query time improvement in percentage
	StorageOverheadMB    float64 // Additional or reduced storage impact
	Complexity           string  // qualitative value: low, medium, high
	Risk                 string  // qualitative value: low, medium, high
	AdditionalNotes      string  // optional descriptive text
}

// OptimizationProposal represents a single optimization suggestion made by an agent.
type OptimizationProposal struct {
	ID               int64
	AgentExecutionID int64
	ProposalType     values.ProposalType
	SQLCommands      []string
	Rationale        string
	EstimatedImpact  EstimatedImpact
	CreatedAt        time.Time
}

// Validate checks whether the OptimizationProposal satisfies domain constraints.
func (p *OptimizationProposal) Validate() error {
	if p.AgentExecutionID <= 0 {
		return errors.New("agent_execution_id must be positive")
	}

	switch p.ProposalType {
	case values.ProposalIndex, values.ProposalPartialIndex, values.ProposalCompositeIndex, values.ProposalMaterializedView, values.ProposalPartitioning, values.ProposalDenormalization, values.ProposalQueryRewrite:
		// valid type
	default:
		return errors.New("invalid proposal type")
	}

	if len(p.SQLCommands) == 0 {
		return errors.New("sql_commands cannot be empty")
	}

	for _, cmd := range p.SQLCommands {
		if strings.TrimSpace(cmd) == "" {
			return errors.New("sql_commands cannot contain empty statements")
		}
	}

	if strings.TrimSpace(p.Rationale) == "" {
		return errors.New("rationale cannot be empty")
	}

	if p.EstimatedImpact.QueryTimeImprovement == 0 {
		return errors.New("estimated performance gain must be non-zero")
	}

	if p.EstimatedImpact.Risk == "" {
		return errors.New("risk level must be specified")
	}
	if !isValidLevel(p.EstimatedImpact.Risk) {
		return errors.New("invalid risk level; must be low, medium, or high")
	}

	if p.EstimatedImpact.Complexity == "" {
		return errors.New("complexity level must be specified")
	}
	if !isValidLevel(p.EstimatedImpact.Complexity) {
		return errors.New("invalid complexity level; must be low, medium, or high")
	}

	return nil
}

func isValidLevel(v string) bool {
	switch strings.ToLower(v) {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

// Summary returns a short human-readable description for logs or auditing.
func (p *OptimizationProposal) Summary() string {
	return strings.TrimSpace(
		p.ProposalType.String() + " → " + p.Rationale,
	)
}

