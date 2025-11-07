package entities

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

// ConsensusDecision represents the final decision made by the consensus engine.
// It links to one task and stores the selected winning proposal and its scoring breakdown.
type ConsensusDecision struct {
	ID                int64
	TaskID            int64
	WinningProposalID *int64 // nullable
	AllScores         map[values.AgentType]ProposalScore
	DecisionRationale string
	AppliedToMain     bool
	CreatedAt         time.Time
}

// ProposalScore holds individual proposal scoring results across multiple criteria.
type ProposalScore struct {
	ProposalID         int64
	Performance        float64 // 0–100 scale
	Storage            float64 // 0–100 scale
	Complexity         float64 // 0–100 scale
	Risk               float64 // 0–100 scale
	WeightedTotal      float64 // computed using ScoringCriteria
	Rank               int
	ImprovementPct     float64
	StorageOverheadMB  float64
}

// ScoringCriteria defines configurable weights for scoring categories.
type ScoringCriteria struct {
	PerformanceWeight float64
	StorageWeight     float64
	ComplexityWeight  float64
	RiskWeight        float64
}

// Validate ensures weights sum to 1.0 (± tolerance) and are within valid range.
func (s ScoringCriteria) Validate() error {
	sum := s.PerformanceWeight + s.StorageWeight + s.ComplexityWeight + s.RiskWeight
	if math.Abs(sum-1.0) > 0.0001 {
		return fmt.Errorf("invalid scoring criteria: weights must sum to 1.0, got %.4f", sum)
	}
	weights := []float64{s.PerformanceWeight, s.StorageWeight, s.ComplexityWeight, s.RiskWeight}
	for _, w := range weights {
		if w < 0 || w > 1 {
			return errors.New("weights must be between 0.0 and 1.0")
		}
	}
	return nil
}

// CalculateWeightedTotal computes the overall score for a given proposal.
func (s ScoringCriteria) CalculateWeightedTotal(score ProposalScore) float64 {
	weighted := (score.Performance * s.PerformanceWeight) +
		(score.Storage * s.StorageWeight) +
		(score.Complexity * s.ComplexityWeight) +
		(score.Risk * s.RiskWeight)
	return math.Round(weighted*100) / 100 // 2 decimal precision
}

// ApplyScores updates WeightedTotal for all proposal scores in a decision.
func (cd *ConsensusDecision) ApplyScores(criteria ScoringCriteria) error {
	if err := criteria.Validate(); err != nil {
		return err
	}
	for agentType, s := range cd.AllScores {
		s.WeightedTotal = criteria.CalculateWeightedTotal(s)
		cd.AllScores[agentType] = s
	}
	return nil
}
