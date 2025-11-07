package entities

import "github.com/tuusuario/afs-challenge/internal/domain/values"

const (
    ProposalTypeIndex            = values.ProposalIndex
    ProposalTypePartialIndex     = values.ProposalPartialIndex
    ProposalTypeCompositeIndex   = values.ProposalCompositeIndex
    ProposalTypeMaterializedView = values.ProposalMaterializedView
    ProposalTypePartitioning     = values.ProposalPartitioning
    ProposalTypeDenormalization  = values.ProposalDenormalization
    ProposalTypeQueryRewrite     = values.ProposalQueryRewrite
)
