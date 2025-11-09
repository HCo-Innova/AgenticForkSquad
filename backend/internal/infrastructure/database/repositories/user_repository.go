package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/tuusuario/afs-challenge/internal/domain/entities"
	"time"
)

type PostgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(email, passwordHash, role, fullName string) (*entities.User, error) {
	query := `
		INSERT INTO users (email, password_hash, role, full_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, email, password_hash, role, full_name, created_at, updated_at, last_login, is_active
	`
	var user entities.User
	err := r.db.QueryRowx(query, email, passwordHash, role, fullName).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) FindByEmail(email string) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, role, full_name, created_at, updated_at, last_login, is_active
		FROM users
		WHERE email = $1 AND is_active = TRUE
	`
	var user entities.User
	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) FindByID(id int) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, role, full_name, created_at, updated_at, last_login, is_active
		FROM users
		WHERE id = $1 AND is_active = TRUE
	`
	var user entities.User
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) UpdateLastLogin(id int) error {
	query := `UPDATE users SET last_login = $1 WHERE id = $2`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

func (r *PostgresUserRepository) UpdatePassword(id int, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, passwordHash, id)
	return err
}

func (r *PostgresUserRepository) List() ([]*entities.User, error) {
	query := `
		SELECT id, email, password_hash, role, full_name, created_at, updated_at, last_login, is_active
		FROM users
		WHERE is_active = TRUE
		ORDER BY created_at DESC
	`
	var users []*entities.User
	err := r.db.Select(&users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}
