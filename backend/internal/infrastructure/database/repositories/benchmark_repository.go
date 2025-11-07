package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
)

type PostgresBenchmarkRepository struct{ db *sqlx.DB }

func NewPostgresBenchmarkRepository(db *sqlx.DB) domainif.BenchmarkRepository {
	return &PostgresBenchmarkRepository{db: db}
}

type benchmarkRow struct {
	ID              int64           `db:"id"`
	ProposalID      int64           `db:"proposal_id"`
	QueryName       string          `db:"query_name"`
	QueryExecuted   string          `db:"query_executed"`
	ExecutionTimeMS float64         `db:"execution_time_ms"`
	RowsReturned    sql.NullInt64   `db:"rows_returned"`
	ExplainPlan     json.RawMessage `db:"explain_plan"`
	StorageImpactMB sql.NullFloat64 `db:"storage_impact_mb"`
	CreatedAt       time.Time       `db:"created_at"`
}

func (r *PostgresBenchmarkRepository) Create(ctx context.Context, b *entities.BenchmarkResult) error {
	if r.db == nil { return errors.New("nil db") }
	if err := b.Validate(); err != nil { return err }
	plan, err := json.Marshal(b.ExplainPlan)
	if err != nil { return err }
	var rows sql.NullInt64
	if b.RowsReturned != 0 { rows = sql.NullInt64{Int64: b.RowsReturned, Valid: true} }
	var storage sql.NullFloat64
	if b.StorageImpactMB != 0 { storage = sql.NullFloat64{Float64: b.StorageImpactMB, Valid: true} }
	q := `INSERT INTO benchmark_results (proposal_id, query_name, query_executed, execution_time_ms, rows_returned, explain_plan, storage_impact_mb, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7, COALESCE($8, NOW()))
		RETURNING id, created_at`
	err = r.db.QueryRowxContext(ctx, q,
		b.ProposalID,
		string(b.QueryName),
		b.QueryExecuted,
		b.ExecutionTimeMS,
		rows,
		json.RawMessage(plan),
		storage,
		b.CreatedAt,
	).Scan(&b.ID, &b.CreatedAt)
	return err
}

func (r *PostgresBenchmarkRepository) GetByProposalID(ctx context.Context, proposalID int) ([]*entities.BenchmarkResult, error) {
	if r.db == nil { return nil, errors.New("nil db") }
	q := `SELECT id, proposal_id, query_name, query_executed, execution_time_ms, rows_returned, explain_plan, storage_impact_mb, created_at
		FROM benchmark_results WHERE proposal_id=$1 ORDER BY id`
	rows := []benchmarkRow{}
	if err := r.db.SelectContext(ctx, &rows, q, proposalID); err != nil { return nil, err }
	out := make([]*entities.BenchmarkResult, 0, len(rows))
	for _, rr := range rows {
		ent, err := rr.toEntity(); if err != nil { return nil, err }
		out = append(out, ent)
	}
	return out, nil
}

func (br benchmarkRow) toEntity() (*entities.BenchmarkResult, error) {
	var plan entities.ExplainPlan
	if len(br.ExplainPlan) > 0 {
		if err := json.Unmarshal(br.ExplainPlan, &plan); err != nil { return nil, err }
	}
	var rows int64
	if br.RowsReturned.Valid { rows = br.RowsReturned.Int64 }
	var storage float64
	if br.StorageImpactMB.Valid { storage = br.StorageImpactMB.Float64 }
	return &entities.BenchmarkResult{
		ID:              br.ID,
		ProposalID:      br.ProposalID,
		QueryName:       entities.BenchmarkQueryName(br.QueryName),
		QueryExecuted:   br.QueryExecuted,
		ExecutionTimeMS: br.ExecutionTimeMS,
		RowsReturned:    rows,
		ExplainPlan:     plan,
		StorageImpactMB: storage,
		CreatedAt:       br.CreatedAt,
	}, nil
}
