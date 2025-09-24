package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user of the EventBooker system.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}
