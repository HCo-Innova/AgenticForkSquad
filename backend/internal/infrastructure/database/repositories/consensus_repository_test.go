package repositories

import (
	"context"
	"testing"
	"time"
	_ "github.com/lib/pq"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/values"
)

func TestConsensusRepository(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()
	ctx := context.Background()

	repo := NewPostgresConsensusRepository(db)

	// prerequisites: task and optional winning proposal
	taskID := insertTaskHelper(t, db)
	execID := insertAgentExecHelper(t, db, taskID)
	propID := insertProposalHelper(t, db, execID)

	// Create decision without winner initially
	dec := &entities.ConsensusDecision{
		TaskID: taskID,
		AllScores: map[values.AgentType]entities.ProposalScore{
			values.AgentCerebro: {ProposalID: propID, Performance: 90, Storage: 80, Complexity: 85, Risk: 88},
		},
		DecisionRationale: "initial",
		AppliedToMain:     false,
		CreatedAt:         time.Now().UTC(),
	}
	if err := repo.Create(ctx, dec); err != nil { t.Fatalf("create err: %v", err) }

	// Read by task
	got, err := repo.GetByTaskID(ctx, int(taskID))
	if err != nil { t.Fatalf("get err: %v", err) }
	if got.TaskID != taskID { t.Fatalf("taskID mismatch") }
	if len(got.AllScores) == 0 { t.Fatalf("expected scores") }

	// Update: set winner and applied flag
	got.DecisionRationale = "final"
	got.AppliedToMain = true
	got.WinningProposalID = &propID
	if err := repo.Update(ctx, got); err != nil { t.Fatalf("update err: %v", err) }
}
