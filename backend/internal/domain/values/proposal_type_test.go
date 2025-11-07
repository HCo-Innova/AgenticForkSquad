package values

import "testing"

func TestProposalTypesExist(t *testing.T) {
	types := []ProposalType{
		ProposalIndex,
		ProposalPartialIndex,
		ProposalCompositeIndex,
		ProposalMaterializedView,
		ProposalPartitioning,
		ProposalDenormalization,
		ProposalQueryRewrite,
	}
	for _, p := range types {
		if p == "" {
			t.Error("found empty proposal type")
		}
	}
}
