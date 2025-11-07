package values

type ProposalType string

const (
	ProposalIndex           ProposalType = "index"
	ProposalPartialIndex    ProposalType = "partial_index"
	ProposalCompositeIndex  ProposalType = "composite_index"
	ProposalMaterializedView ProposalType = "materialized_view"
	ProposalPartitioning    ProposalType = "partitioning"
	ProposalDenormalization ProposalType = "denormalization"
	ProposalQueryRewrite    ProposalType = "query_rewrite"
)

// String converts ProposalType to string safely.
func (pt ProposalType) String() string {
	return string(pt)
}