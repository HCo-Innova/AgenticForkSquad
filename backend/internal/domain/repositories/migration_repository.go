package repositories

import "github.com/tuusuario/afs-challenge/internal/domain/entities"

type MigrationRepository interface {
	Create(migration *entities.Migration) error
	FindByID(id string) (*entities.Migration, error)
	FindAll() ([]*entities.Migration, error)
	Update(migration *entities.Migration) error
	Delete(id string) error
}
