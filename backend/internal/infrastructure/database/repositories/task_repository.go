package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	domainif "github.com/tuusuario/afs-challenge/internal/domain/interfaces"
)

// PostgresTaskRepository implements TaskRepository using PostgreSQL (sqlx).
type PostgresTaskRepository struct {
	db *sqlx.DB
}

func NewPostgresTaskRepository(db *sqlx.DB) domainif.TaskRepository {
	return &PostgresTaskRepository{db: db}
}

type taskRow struct {
	ID          int64          `db:"id"`
	Type        string         `db:"type"`
	Description sql.NullString `db:"description"`
	TargetQuery string         `db:"target_query"`
	Status      string         `db:"status"`
	CreatedAt   time.Time      `db:"created_at"`
	CompletedAt sql.NullTime   `db:"completed_at"`
	Metadata    []byte         `db:"metadata"`
}

func toRow(t *entities.Task) (taskRow, error) {
	var metaBytes []byte
	var err error
	if t.Metadata != nil {
		metaBytes, err = json.Marshal(t.Metadata)
		if err != nil {
			return taskRow{}, err
		}
	}
	var desc sql.NullString
	if strings.TrimSpace(t.Description) != "" {
		desc = sql.NullString{String: t.Description, Valid: true}
	}
	var completed sql.NullTime
	if t.CompletedAt != nil {
		completed = sql.NullTime{Time: *t.CompletedAt, Valid: true}
	}
	return taskRow{
		ID:          t.ID,
		Type:        string(t.Type),
		Description: desc,
		TargetQuery: t.TargetQuery,
		Status:      string(t.Status),
		CreatedAt:   t.CreatedAt,
		CompletedAt: completed,
		Metadata:    metaBytes,
	}, nil
}

func fromRow(r taskRow) (*entities.Task, error) {
	var meta map[string]interface{}
	if len(r.Metadata) > 0 {
		if err := json.Unmarshal(r.Metadata, &meta); err != nil {
			return nil, err
		}
	}
	var completed *time.Time
	if r.CompletedAt.Valid {
		completed = &r.CompletedAt.Time
	}
	return &entities.Task{
		ID:          r.ID,
		Type:        entities.TaskType(r.Type),
		Description: r.Description.String,
		TargetQuery: r.TargetQuery,
		Status:      entities.TaskStatus(r.Status),
		CreatedAt:   r.CreatedAt,
		CompletedAt: completed,
		Metadata:    meta,
	}, nil
}

// Create inserts a new task and sets ID/CreatedAt.
func (r *PostgresTaskRepository) Create(ctx context.Context, task *entities.Task) error {
	if r == nil || r.db == nil {
		return errors.New("nil repository or db")
	}
	if err := task.Validate(); err != nil {
		return err
	}
	row, err := toRow(task)
	if err != nil { return err }
	// Use NOW() if CreatedAt is zero
	query := `INSERT INTO tasks (type, description, target_query, status, created_at, completed_at, metadata)
		VALUES ($1,$2,$3,$4,COALESCE($5, NOW()), $6, $7)
		RETURNING id, created_at`
	var createdAt time.Time
	err = r.db.QueryRowxContext(ctx, query,
		row.Type,
		row.Description,
		row.TargetQuery,
		row.Status,
		nullTimeFrom(row.CreatedAt),
		row.CompletedAt,
		jsonRawOrNull(row.Metadata),
	).Scan(&task.ID, &createdAt)
	if err != nil { return err }
	task.CreatedAt = createdAt
	return nil
}

// GetByID fetches a task by id.
func (r *PostgresTaskRepository) GetByID(ctx context.Context, id int) (*entities.Task, error) {
	if r == nil || r.db == nil { return nil, errors.New("nil repository or db") }
	var tr taskRow
	query := `SELECT id, type, description, target_query, status, created_at, completed_at, COALESCE(metadata, '{}'::jsonb) as metadata
		FROM tasks WHERE id=$1`
	if err := r.db.GetContext(ctx, &tr, query, id); err != nil {
		return nil, err
	}
	return fromRow(tr)
}

// List returns tasks filtered by optional fields.
func (r *PostgresTaskRepository) List(ctx context.Context, filters entities.TaskFilters) ([]*entities.Task, error) {
	if r == nil || r.db == nil { return nil, errors.New("nil repository or db") }
	conds := []string{}
	args := []interface{}{}
	if strings.TrimSpace(filters.Status) != "" {
		args = append(args, filters.Status)
		conds = append(conds, fmt.Sprintf("status=$%d", len(args)))
	}
	if strings.TrimSpace(filters.Type) != "" {
		args = append(args, filters.Type)
		conds = append(conds, fmt.Sprintf("type=$%d", len(args)))
	}
	if strings.TrimSpace(filters.CreatedAfter) != "" {
		args = append(args, filters.CreatedAfter)
		conds = append(conds, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if strings.TrimSpace(filters.CreatedBefore) != "" {
		args = append(args, filters.CreatedBefore)
		conds = append(conds, fmt.Sprintf("created_at <= $%d", len(args)))
	}
	base := `SELECT id, type, description, target_query, status, created_at, completed_at, COALESCE(metadata, '{}'::jsonb) as metadata FROM tasks`
	if len(conds) > 0 {
		base += " WHERE " + strings.Join(conds, " AND ")
	}
	base += " ORDER BY created_at DESC"
	rows := []taskRow{}
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*entities.Task, 0, len(rows))
	for _, rr := range rows {
		ent, err := fromRow(rr)
		if err != nil { return nil, err }
		out = append(out, ent)
	}
	return out, nil
}

// Update updates fields of an existing task.
func (r *PostgresTaskRepository) Update(ctx context.Context, task *entities.Task) error {
	if r == nil || r.db == nil { return errors.New("nil repository or db") }
	if task.ID == 0 { return errors.New("missing id") }
	if err := task.Validate(); err != nil { return err }
	row, err := toRow(task)
	if err != nil { return err }
	query := `UPDATE tasks SET type=$1, description=$2, target_query=$3, status=$4, completed_at=$5, metadata=$6 WHERE id=$7`
	res, err := r.db.ExecContext(ctx, query,
		row.Type,
		row.Description,
		row.TargetQuery,
		row.Status,
		row.CompletedAt,
		jsonRawOrNull(row.Metadata),
		row.ID,
	)
	if err != nil { return err }
	affected, _ := res.RowsAffected()
	if affected == 0 { return sql.ErrNoRows }
	return nil
}

// Delete removes a task by ID.
func (r *PostgresTaskRepository) Delete(ctx context.Context, id int) error {
	if r == nil || r.db == nil { return errors.New("nil repository or db") }
	if id <= 0 { return errors.New("invalid id") }
	query := `DELETE FROM tasks WHERE id=$1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil { return err }
	affected, _ := res.RowsAffected()
	if affected == 0 { return sql.ErrNoRows }
	return nil
}

// Helpers
func nullTimeFrom(t time.Time) sql.NullTime {
	if t.IsZero() { return sql.NullTime{} }
	return sql.NullTime{Time: t, Valid: true}
}

func jsonRawOrNull(b []byte) interface{} {
	if len(b) == 0 { return nil }
	return json.RawMessage(b)
}
