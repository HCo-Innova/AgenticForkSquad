package entities

import (
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

func TestScoringCriteriaValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   ScoringCriteria
		wantErr bool
	}{
		{"ValidWeights", ScoringCriteria{PerformanceWeight: 0.5, StorageWeight: 0.2, ComplexityWeight: 0.2, RiskWeight: 0.1}, false},
		{"SumNotOne", ScoringCriteria{PerformanceWeight: 0.5, StorageWeight: 0.2, ComplexityWeight: 0.2, RiskWeight: 0.2}, true},
		{"NegativeWeight", ScoringCriteria{PerformanceWeight: -0.1, StorageWeight: 0.5, ComplexityWeight: 0.3, RiskWeight: 0.3}, true},
	}

	for _, tt := range tests {
		err := tt.input.Validate()
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error=%v, got=%v", tt.name, tt.wantErr, err != nil)
		}
	}
}

func TestCalculateWeightedTotal(t *testing.T) {
	criteria := ScoringCriteria{
		PerformanceWeight: 0.5,
		StorageWeight:     0.2,
		ComplexityWeight:  0.2,
		RiskWeight:        0.1,
	}

	score := ProposalScore{
		Performance: 95,
		Storage:     80,
		Complexity:  85,
		Risk:        90,
	}

	got := criteria.CalculateWeightedTotal(score)
	want := 89.5 // (95*0.5)+(80*0.2)+(85*0.2)+(90*0.1)=47.5+16+17+9=89.5

	if got != want {
		t.Errorf("Expected %.2f, got %.2f", want, got)
	}
}

func TestApplyScores(t *testing.T) {
	cd := &ConsensusDecision{
		AllScores: map[values.AgentType]ProposalScore{
			values.AgentCerebro:  {ProposalID: 1, Performance: 100, Storage: 90, Complexity: 80, Risk: 70},
			values.AgentOperativo: {ProposalID: 2, Performance: 80, Storage: 100, Complexity: 70, Risk: 60},
		},
	}

	criteria := ScoringCriteria{
		PerformanceWeight: 0.5,
		StorageWeight:     0.2,
		ComplexityWeight:  0.2,
		RiskWeight:        0.1,
	}
	err := cd.ApplyScores(criteria)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if cd.AllScores[values.AgentCerebro].WeightedTotal == 0 || cd.AllScores[values.AgentOperativo].WeightedTotal == 0 {
		t.Errorf("weighted totals not computed properly")
	}
}
