package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
)

func connectTestDB(t *testing.T) *sqlx.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback to POSTGRES_* if available
		host := getenv("POSTGRES_HOST", "postgres")
		port := getenv("POSTGRES_PORT", "5432")
		db := os.Getenv("POSTGRES_DB")
		user := os.Getenv("POSTGRES_USER")
		pass := os.Getenv("POSTGRES_PASSWORD")
		if db != "" && user != "" && pass != "" {
			dsn = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + db + "?sslmode=disable"
		}
	}
	if dsn == "" {
		t.Skip("DATABASE_URL/POSTGRES_* not set; skipping DB integration test")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Skipf("cannot connect to db: %v", err)
	}
	return db
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

func TestTaskRepository(t *testing.T) {
	db := connectTestDB(t)
	defer db.Close()
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	// Create
	now := time.Now().UTC()
	task := &entities.Task{
		Type:        entities.TaskTypeQueryOptimization,
		Description: "integration test",
		TargetQuery: "SELECT 1",
		Status:      entities.TaskStatusPending,
		CreatedAt:   now,
		Metadata:    map[string]interface{}{"k":"v"},
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create err: %v", err)
	}
	if task.ID == 0 { t.Fatalf("expected ID assigned") }

	// GetByID
	read, err := repo.GetByID(ctx, int(task.ID))
	if err != nil { t.Fatalf("get err: %v", err) }
	if read.TargetQuery != task.TargetQuery { t.Fatalf("mismatch query") }

	// List with filter
	list, err := repo.List(ctx, entities.TaskFilters{Status: string(entities.TaskStatusPending)})
	if err != nil { t.Fatalf("list err: %v", err) }
	if len(list) == 0 { t.Fatalf("expected at least one row") }

	// Update
	read.Status = entities.TaskStatusInProgress
	if err := repo.Update(ctx, read); err != nil { t.Fatalf("update err: %v", err) }
	again, err := repo.GetByID(ctx, int(task.ID))
	if err != nil { t.Fatalf("get after update err: %v", err) }
	if again.Status != entities.TaskStatusInProgress { t.Fatalf("status not updated") }
}
