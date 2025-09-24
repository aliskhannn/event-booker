package model

import (
	"time"

	"github.com/google/uuid"
)

// Booking represents a reservation for a seat at an event.
type Booking struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
