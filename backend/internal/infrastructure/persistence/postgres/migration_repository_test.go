package postgres

import "testing"

func TestNewPostgresMigrationRepository(t *testing.T) {
	repo := NewPostgresMigrationRepository()
	if repo == nil {
		t.Fatalf("expected non-nil repo")
	}
}

func TestMigrationRepository_TODOsNoPanic(t *testing.T) {
	repo := NewPostgresMigrationRepository().(*PostgresMigrationRepository)
	if err := repo.Create(nil); err != nil { t.Fatalf("unexpected error: %v", err) }
	if _, err := repo.FindByID("id"); err != nil { t.Fatalf("unexpected error: %v", err) }
	if _, err := repo.FindAll(); err != nil { t.Fatalf("unexpected error: %v", err) }
	if err := repo.Update(nil); err != nil { t.Fatalf("unexpected error: %v", err) }
	if err := repo.Delete("id"); err != nil { t.Fatalf("unexpected error: %v", err) }
}
