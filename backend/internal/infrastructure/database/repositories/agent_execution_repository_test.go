package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

func TestAgentExecutionRepository(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()
	ctx := context.Background()
	// ensure a task exists
	taskID := insertTaskHelper(t, db)
	repo := NewPostgresAgentExecutionRepository(db)

	exec := &entities.AgentExecution{
		TaskID:    int64(taskID),
		AgentType: values.AgentOperativo,
		ForkID:    "afs-fork-operativo-task1-123",
		Status:    entities.ExecutionRunning,
		StartedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, exec); err != nil { t.Fatalf("create err: %v", err) }
	got, err := repo.GetByID(ctx, int(exec.ID))
	if err != nil { t.Fatalf("get err: %v", err) }
	if got.TaskID != exec.TaskID { t.Fatalf("taskID mismatch") }

	list, err := repo.GetByTaskID(ctx, int(exec.TaskID))
	if err != nil || len(list) == 0 { t.Fatalf("list err: %v n=%d", err, len(list)) }

	got.Status = entities.ExecutionCompleted
	now := time.Now().UTC()
	got.CompletedAt = &now
	if err := repo.Update(ctx, got); err != nil { t.Fatalf("update err: %v", err) }
}

// helper: insert a minimal task row and return ID
func insertTaskHelper(t *testing.T, db *sqlx.DB) int64 {
	var id int64
	q := `INSERT INTO tasks (type, target_query, status, created_at) VALUES ($1,$2,'pending', NOW()) RETURNING id`
	if err := db.QueryRowx(q, "query_optimization", "SELECT 1").Scan(&id); err != nil {
		t.Skipf("cannot insert task (migrations may be missing): %v", err)
	}
	return id
}
