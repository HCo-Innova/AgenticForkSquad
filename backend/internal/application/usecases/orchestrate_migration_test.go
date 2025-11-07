package usecases

import (
	"errors"
	"testing"

	"github.com/tuusuario/afs-challenge/internal/domain/entities"
)

type fakeMigrationRepo struct{
	created []*entities.Migration
	err error
}

func (f *fakeMigrationRepo) Create(m *entities.Migration) error { f.created = append(f.created, m); return f.err }
func (f *fakeMigrationRepo) FindByID(id string) (*entities.Migration, error) { return nil, f.err }
func (f *fakeMigrationRepo) FindAll() ([]*entities.Migration, error) { return nil, f.err }
func (f *fakeMigrationRepo) Update(m *entities.Migration) error { return f.err }
func (f *fakeMigrationRepo) Delete(id string) error { return f.err }

func TestNewOrchestrateMigrationUseCase(t *testing.T) {
	repo := &fakeMigrationRepo{}
	uc := NewOrchestrateMigrationUseCase(repo)
	if uc == nil {
		t.Fatalf("expected non-nil use case")
	}
}

func TestExecuteReturnsNilForNow(t *testing.T) {
	repo := &fakeMigrationRepo{}
	uc := NewOrchestrateMigrationUseCase(repo)
	if err := uc.Execute("init", "safe"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestExecuteWithRepoErrorPathDoesNotPropagateForNow(t *testing.T) {
	repo := &fakeMigrationRepo{err: errors.New("boom")}
	uc := NewOrchestrateMigrationUseCase(repo)
	if err := uc.Execute("init", "safe"); err != nil {
		t.Fatalf("expected nil error (no orchestration yet), got %v", err)
	}
}
