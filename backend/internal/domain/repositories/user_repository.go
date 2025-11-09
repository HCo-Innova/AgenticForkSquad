package repositories

import "github.com/tuusuario/afs-challenge/internal/domain/entities"

type UserRepository interface {
	Create(email, passwordHash, role, fullName string) (*entities.User, error)
	FindByEmail(email string) (*entities.User, error)
	FindByID(id int) (*entities.User, error)
	UpdateLastLogin(id int) error
	UpdatePassword(id int, passwordHash string) error
	List() ([]*entities.User, error)
}
