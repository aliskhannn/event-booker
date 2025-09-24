package model

import (
	"time"

	"github.com/google/uuid"
)

// Event represents an event that users can book seats for.
type Event struct {
	ID             uuid.UUID     `json:"id"`
	Title          string        `json:"title"`
	Date           time.Time     `json:"date"`
	TotalSeats     int           `json:"total_seats"`
	AvailableSeats int           `json:"available_seats"`
	BookingTTL     time.Duration `json:"booking_ttl"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}
