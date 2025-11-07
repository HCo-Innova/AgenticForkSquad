package entities

import "time"

type Migration struct {
	ID          string
	Name        string
	Strategy    string
	Status      string
	StartedAt   time.Time
	CompletedAt *time.Time
}
