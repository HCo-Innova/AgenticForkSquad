package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

type PostgresOptimizationRepository struct{ db *sqlx.DB }

func NewPostgresOptimizationRepository(db *sqlx.DB) domainif.OptimizationRepository {
	return &PostgresOptimizationRepository{db: db}
}

type proposalRow struct {
	ID               int64           `db:"id"`
	AgentExecutionID int64           `db:"agent_execution_id"`
	ProposalType     string          `db:"proposal_type"`
	SQLCommands      pq.StringArray  `db:"sql_commands"`
	Rationale        sql.NullString  `db:"rationale"`
	EstimatedImpact  json.RawMessage `db:"estimated_impact"`
	CreatedAt        time.Time       `db:"created_at"`
}

func (r *PostgresOptimizationRepository) Create(ctx context.Context, p *entities.OptimizationProposal) error {
	if r.db == nil { return errors.New("nil db") }
	if err := p.Validate(); err != nil { return err }
	imp, err := json.Marshal(p.EstimatedImpact)
	if err != nil { return err }
	q := `INSERT INTO optimization_proposals (agent_execution_id, proposal_type, sql_commands, rationale, estimated_impact, created_at)
		VALUES ($1,$2,$3,$4,$5, COALESCE($6, NOW()))
		RETURNING id, created_at`
	var rationale sql.NullString
	if p.Rationale != "" { rationale = sql.NullString{String: p.Rationale, Valid: true} }
	err = r.db.QueryRowxContext(ctx, q,
		p.AgentExecutionID,
		string(p.ProposalType),
		pq.Array(p.SQLCommands),
		rationale,
		json.RawMessage(imp),
		p.CreatedAt,
	).Scan(&p.ID, &p.CreatedAt)
	return err
}

func (r *PostgresOptimizationRepository) GetByID(ctx context.Context, id int) (*entities.OptimizationProposal, error) {
	if r.db == nil { return nil, errors.New("nil db") }
	q := `SELECT id, agent_execution_id, proposal_type, sql_commands, rationale, estimated_impact, created_at FROM optimization_proposals WHERE id=$1`
	var pr proposalRow
	if err := r.db.GetContext(ctx, &pr, q, id); err != nil { return nil, err }
	return pr.toEntity()
}

func (r *PostgresOptimizationRepository) GetByAgentExecutionID(ctx context.Context, execID int) ([]*entities.OptimizationProposal, error) {
	if r.db == nil { return nil, errors.New("nil db") }
	q := `SELECT id, agent_execution_id, proposal_type, sql_commands, rationale, estimated_impact, created_at FROM optimization_proposals WHERE agent_execution_id=$1 ORDER BY id`
	rows := []proposalRow{}
	if err := r.db.SelectContext(ctx, &rows, q, execID); err != nil { return nil, err }
	out := make([]*entities.OptimizationProposal, 0, len(rows))
	for _, rr := range rows {
		ent, err := rr.toEntity(); if err != nil { return nil, err }
		out = append(out, ent)
	}
	return out, nil
}

func (r *PostgresOptimizationRepository) Update(ctx context.Context, p *entities.OptimizationProposal) error {
	if r.db == nil { return errors.New("nil db") }
	if p.ID == 0 { return errors.New("missing id") }
	if err := p.Validate(); err != nil { return err }
	imp, err := json.Marshal(p.EstimatedImpact)
	if err != nil { return err }
	var rationale sql.NullString
	if p.Rationale != "" { rationale = sql.NullString{String: p.Rationale, Valid: true} }
	q := `UPDATE optimization_proposals SET proposal_type=$1, sql_commands=$2, rationale=$3, estimated_impact=$4 WHERE id=$5`
	_, err = r.db.ExecContext(ctx, q,
		string(p.ProposalType),
		pq.Array(p.SQLCommands),
		rationale,
		json.RawMessage(imp),
		p.ID,
	)
	return err
}

func (pr proposalRow) toEntity() (*entities.OptimizationProposal, error) {
	var imp entities.EstimatedImpact
	if len(pr.EstimatedImpact) > 0 {
		if err := json.Unmarshal(pr.EstimatedImpact, &imp); err != nil { return nil, err }
	}
	return &entities.OptimizationProposal{
		ID:               pr.ID,
		AgentExecutionID: pr.AgentExecutionID,
		ProposalType:     values.ProposalType(pr.ProposalType),
		SQLCommands:      []string(pr.SQLCommands),
		Rationale:        pr.Rationale.String,
		EstimatedImpact:  imp,
		CreatedAt:        pr.CreatedAt,
	}, nil
}

// kept for compatibility if needed elsewhere
func pqStringArray(a []string) interface{} { return pq.Array(a) }
