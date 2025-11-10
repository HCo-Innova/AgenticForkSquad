package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

type PostgresAgentExecutionRepository struct {
	db *sqlx.DB
}

func NewPostgresAgentExecutionRepository(db *sqlx.DB) domainif.AgentExecutionRepository {
	return &PostgresAgentExecutionRepository{db: db}
}

type agentExecRow struct {
	ID          int64        `db:"id"`
	TaskID      int64        `db:"task_id"`
	AgentType   string       `db:"agent_type"`
	ForkID      sql.NullString `db:"fork_id"`
	Status      string       `db:"status"`
	StartedAt   time.Time    `db:"started_at"`
	CompletedAt sql.NullTime `db:"completed_at"`
	ErrorMsg    sql.NullString `db:"error_message"`
}

func (r *PostgresAgentExecutionRepository) Create(ctx context.Context, exec *entities.AgentExecution) error {
	if r.db == nil { return errors.New("nil db") }
	if exec.Status == "" { exec.Status = entities.ExecutionRunning }
	q := `INSERT INTO agent_executions (task_id, agent_type, fork_id, status, started_at, completed_at, error_message)
		VALUES ($1,$2,$3,COALESCE($4,'running'), COALESCE($5,NOW()), $6, $7)
		RETURNING id, started_at`
	var started time.Time
	var fork sql.NullString
	if exec.ForkID != "" { fork = sql.NullString{String: exec.ForkID, Valid: true} }
	var completed sql.NullTime
	if exec.CompletedAt != nil { completed = sql.NullTime{Time: *exec.CompletedAt, Valid: true} }
	var errMsg sql.NullString
	if exec.ErrorMsg != "" { errMsg = sql.NullString{String: exec.ErrorMsg, Valid: true} }
	err := r.db.QueryRowxContext(ctx, q,
		exec.TaskID,
		string(exec.AgentType),
		fork,
		string(exec.Status),
		exec.StartedAt,
		completed,
		errMsg,
	).Scan(&exec.ID, &started)
	if err != nil { return err }
	exec.StartedAt = started
	return nil
}

func (r *PostgresAgentExecutionRepository) GetByID(ctx context.Context, id int) (*entities.AgentExecution, error) {
	if r.db == nil { return nil, errors.New("nil db") }
	q := `SELECT id, task_id, agent_type, fork_id, status, started_at, completed_at, error_message FROM agent_executions WHERE id=$1`
	var rw agentExecRow
	if err := r.db.GetContext(ctx, &rw, q, id); err != nil { return nil, err }
	return rw.toEntity(), nil
}

func (r *PostgresAgentExecutionRepository) GetByTaskID(ctx context.Context, taskID int) ([]*entities.AgentExecution, error) {
	if r.db == nil { return nil, errors.New("nil db") }
	q := `SELECT id, task_id, agent_type, fork_id, status, started_at, completed_at, error_message FROM agent_executions WHERE task_id=$1 ORDER BY id`
	rows := []agentExecRow{}
	if err := r.db.SelectContext(ctx, &rows, q, taskID); err != nil { return nil, err }
	out := make([]*entities.AgentExecution, 0, len(rows))
	for _, rr := range rows { out = append(out, rr.toEntity()) }
	return out, nil
}

func (r *PostgresAgentExecutionRepository) List(ctx context.Context) ([]*entities.AgentExecution, error) {
	if r.db == nil { return nil, errors.New("nil db") }
	q := `SELECT id, task_id, agent_type, fork_id, status, started_at, completed_at, error_message FROM agent_executions ORDER BY started_at DESC LIMIT 100`
	rows := []agentExecRow{}
	if err := r.db.SelectContext(ctx, &rows, q); err != nil { return nil, err }
	out := make([]*entities.AgentExecution, 0, len(rows))
	for _, rr := range rows { out = append(out, rr.toEntity()) }
	return out, nil
}

func (r *PostgresAgentExecutionRepository) Update(ctx context.Context, exec *entities.AgentExecution) error {
	if r.db == nil { return errors.New("nil db") }
	if exec.ID == 0 { return errors.New("missing id") }
	var fork sql.NullString
	if exec.ForkID != "" { fork = sql.NullString{String: exec.ForkID, Valid: true} }
	var completed sql.NullTime
	if exec.CompletedAt != nil { completed = sql.NullTime{Time: *exec.CompletedAt, Valid: true} }
	var errMsg sql.NullString
	if exec.ErrorMsg != "" { errMsg = sql.NullString{String: exec.ErrorMsg, Valid: true} }
	q := `UPDATE agent_executions SET fork_id=$1, status=$2, completed_at=$3, error_message=$4 WHERE id=$5`
	res, err := r.db.ExecContext(ctx, q, fork, string(exec.Status), completed, errMsg, exec.ID)
	if err != nil { return err }
	a, _ := res.RowsAffected()
	if a == 0 { return sql.ErrNoRows }
	return nil
}

func (r agentExecRow) toEntity() *entities.AgentExecution {
	var completed *time.Time
	if r.CompletedAt.Valid { completed = &r.CompletedAt.Time }
	var fork string
	if r.ForkID.Valid { fork = r.ForkID.String }
	var errMsg string
	if r.ErrorMsg.Valid { errMsg = r.ErrorMsg.String }
	return &entities.AgentExecution{
		ID:          r.ID,
		TaskID:      r.TaskID,
		AgentType:   values.AgentType(r.AgentType),
		ForkID:      fork,
		Status:      entities.ExecutionStatus(r.Status),
		StartedAt:   r.StartedAt,
		CompletedAt: completed,
		ErrorMsg:    errMsg,
	}
}
