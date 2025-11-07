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

func insertAgentExecHelper(t *testing.T, db *sqlx.DB, taskID int64) int64 {
	var id int64
	q := `INSERT INTO agent_executions (task_id, agent_type, status, started_at)
		VALUES ($1,$2,'running', NOW()) RETURNING id`
	if err := db.QueryRowx(q, taskID, "cerebro").Scan(&id); err != nil {
		t.Skipf("cannot insert agent_execution: %v", err)
	}
	return id
}

func TestOptimizationRepository(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskID := insertTaskHelper(t, db)
	execID := insertAgentExecHelper(t, db, taskID)

	repo := NewPostgresOptimizationRepository(db)

	prop := &entities.OptimizationProposal{
		AgentExecutionID: execID,
		ProposalType:     values.ProposalIndex,
		SQLCommands:      []string{"CREATE INDEX idx_test ON orders(user_id);"},
		Rationale:        "index on orders(user_id)",
		EstimatedImpact: entities.EstimatedImpact{
			QueryTimeImprovement: 10,
			StorageOverheadMB:    1.5,
			Complexity:           "low",
			Risk:                 "low",
		},
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, prop); err != nil {
		t.Fatalf("create err: %v", err)
	}
	got, err := repo.GetByID(ctx, int(prop.ID))
	if err != nil { t.Fatalf("get err: %v", err) }
	if got.AgentExecutionID != execID { t.Fatalf("exec id mismatch") }

	list, err := repo.GetByAgentExecutionID(ctx, int(execID))
	if err != nil || len(list) == 0 { t.Fatalf("list err=%v n=%d", err, len(list)) }

	got.Rationale = "updated rationale"
	got.SQLCommands = []string{"CREATE INDEX idx_test2 ON orders(status);"}
	if err := repo.Update(ctx, got); err != nil { t.Fatalf("update err: %v", err) }
}
