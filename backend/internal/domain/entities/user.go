package entities

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	FullName     string    `json:"full_name" db:"full_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleViewer UserRole = "viewer"
)

func (u *User) IsAdmin() bool {
	return u.Role == string(RoleAdmin)
}

func (u *User) CanCreateTask() bool {
	return u.Role == string(RoleAdmin) || u.Role == string(RoleUser)
}

func (u *User) CanViewAllTasks() bool {
	return u.Role == string(RoleAdmin)
}
