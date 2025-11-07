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
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

type PostgresConsensusRepository struct{ db *sqlx.DB }

func NewPostgresConsensusRepository(db *sqlx.DB) domainif.ConsensusRepository {
	return &PostgresConsensusRepository{db: db}
}

type consensusRow struct {
	ID                int64           `db:"id"`
	TaskID            int64           `db:"task_id"`
	WinningProposalID sql.NullInt64   `db:"winning_proposal_id"`
	AllScores         json.RawMessage `db:"all_scores"`
	DecisionRationale sql.NullString  `db:"decision_rationale"`
	AppliedToMain     bool            `db:"applied_to_main"`
	CreatedAt         time.Time       `db:"created_at"`
}

func (r *PostgresConsensusRepository) Create(ctx context.Context, d *entities.ConsensusDecision) error {
	if r.db == nil { return errors.New("nil db") }
	scores := marshalScores(d.AllScores)
	var win sql.NullInt64
	if d.WinningProposalID != nil { win = sql.NullInt64{Int64: *d.WinningProposalID, Valid: true} }
	var rationale sql.NullString
	if d.DecisionRationale != "" { rationale = sql.NullString{String: d.DecisionRationale, Valid: true} }
	q := `INSERT INTO consensus_decisions (task_id, winning_proposal_id, all_scores, decision_rationale, applied_to_main, created_at)
		VALUES ($1,$2,$3,$4, $5, COALESCE($6, NOW()))
		RETURNING id, created_at`
	err := r.db.QueryRowxContext(ctx, q,
		d.TaskID,
		win,
		scores,
		rationale,
		d.AppliedToMain,
		d.CreatedAt,
	).Scan(&d.ID, &d.CreatedAt)
	return err
}

func (r *PostgresConsensusRepository) GetByTaskID(ctx context.Context, taskID int) (*entities.ConsensusDecision, error) {
	if r.db == nil { return nil, errors.New("nil db") }
	q := `SELECT id, task_id, winning_proposal_id, all_scores, decision_rationale, applied_to_main, created_at FROM consensus_decisions WHERE task_id=$1`
	var row consensusRow
	if err := r.db.GetContext(ctx, &row, q, taskID); err != nil { return nil, err }
	return row.toEntity()
}

func (r *PostgresConsensusRepository) Update(ctx context.Context, d *entities.ConsensusDecision) error {
	if r.db == nil { return errors.New("nil db") }
	if d.ID == 0 { return errors.New("missing id") }
	scores := marshalScores(d.AllScores)
	var win sql.NullInt64
	if d.WinningProposalID != nil { win = sql.NullInt64{Int64: *d.WinningProposalID, Valid: true} }
	var rationale sql.NullString
	if d.DecisionRationale != "" { rationale = sql.NullString{String: d.DecisionRationale, Valid: true} }
	q := `UPDATE consensus_decisions SET winning_proposal_id=$1, all_scores=$2, decision_rationale=$3, applied_to_main=$4 WHERE id=$5`
	_, err := r.db.ExecContext(ctx, q, win, scores, rationale, d.AppliedToMain, d.ID)
	return err
}

func (r consensusRow) toEntity() (*entities.ConsensusDecision, error) {
	var win *int64
	if r.WinningProposalID.Valid { v := r.WinningProposalID.Int64; win = &v }
	m, err := unmarshalScores(r.AllScores)
	if err != nil { return nil, err }
	return &entities.ConsensusDecision{
		ID:                r.ID,
		TaskID:            r.TaskID,
		WinningProposalID: win,
		AllScores:         m,
		DecisionRationale: r.DecisionRationale.String,
		AppliedToMain:     r.AppliedToMain,
		CreatedAt:         r.CreatedAt,
	}, nil
}

func marshalScores(m map[values.AgentType]entities.ProposalScore) json.RawMessage {
	if m == nil { return json.RawMessage([]byte(`{}`)) }
	b, _ := json.Marshal(m)
	return json.RawMessage(b)
}

func unmarshalScores(b []byte) (map[values.AgentType]entities.ProposalScore, error) {
	if len(b) == 0 { return map[values.AgentType]entities.ProposalScore{}, nil }
	var tmp map[string]entities.ProposalScore
	if err := json.Unmarshal(b, &tmp); err != nil { return nil, err }
	out := make(map[values.AgentType]entities.ProposalScore, len(tmp))
	for k, v := range tmp { out[values.AgentType(k)] = v }
	return out, nil
}
