package postgres

import (
	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"github.com/tuusuario/afs-challenge/internal/domain/repositories"
)

type PostgresMigrationRepository struct {
	// db *sql.DB
}

func NewPostgresMigrationRepository() repositories.MigrationRepository {
	return &PostgresMigrationRepository{}
}

func (r *PostgresMigrationRepository) Create(migration *entities.Migration) error {
	// TODO: Implement
	return nil
}

func (r *PostgresMigrationRepository) FindByID(id string) (*entities.Migration, error) {
	// TODO: Implement
	return nil, nil
}

func (r *PostgresMigrationRepository) FindAll() ([]*entities.Migration, error) {
	// TODO: Implement
	return nil, nil
}

func (r *PostgresMigrationRepository) Update(migration *entities.Migration) error {
	// TODO: Implement
	return nil
}

func (r *PostgresMigrationRepository) Delete(id string) error {
	// TODO: Implement
	return nil
}
