package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
)

func TestBenchmarkRepository(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()
	ctx := context.Background()

	repo := NewPostgresBenchmarkRepository(db)
	// prerequisite: create task, agent_execution, proposal
	taskID := insertTaskHelper(t, db)
	execID := insertAgentExecHelper(t, db, taskID)
	propID := insertProposalHelper(t, db, execID)

	b := &entities.BenchmarkResult{
		ProposalID:      propID,
		QueryName:       entities.QueryNameBaseline,
		QueryExecuted:   "SELECT 1",
		ExecutionTimeMS: 12.34,
		RowsReturned:    1,
		ExplainPlan: entities.ExplainPlan{PlanType: "Seq Scan"},
		StorageImpactMB: 0,
		CreatedAt:       time.Now().UTC(),
	}
	if err := repo.Create(ctx, b); err != nil { t.Fatalf("create err: %v", err) }

	list, err := repo.GetByProposalID(ctx, int(propID))
	if err != nil || len(list) == 0 { t.Fatalf("list err=%v n=%d", err, len(list)) }
}

// helper: insert a basic proposal row
func insertProposalHelper(t *testing.T, db *sqlx.DB, execID int64) int64 {
	var id int64
	q := `INSERT INTO optimization_proposals (agent_execution_id, proposal_type, sql_commands, estimated_impact, created_at)
		VALUES ($1,'index', ARRAY['SELECT 1']::text[], '{}'::jsonb, NOW()) RETURNING id`
	if err := db.QueryRowx(q, execID).Scan(&id); err != nil {
		t.Skipf("cannot insert proposal: %v", err)
	}
	return id
}
