package event

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/event-booker/internal/model"
)

var (
	ErrEventNotFound                     = errors.New("event not found")
	ErrNoSeatsAvailable                  = errors.New("no seats available")
	ErrBookingNotFoundOrAlreadyConfirmed = errors.New("booking not found or already confirmed")
	ErrBookingNotFoundOrAlreadyCancelled = errors.New("booking not found or already cancelled")
)

// Repository provides methods to interact with events table.
type Repository struct {
	db *dbpg.DB
}

// NewRepository creates a new event repository.
func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

// CreateEvent adds a new event to the database.
func (r *Repository) CreateEvent(ctx context.Context, event *model.Event) (uuid.UUID, error) {
	query := `
		INSERT INTO events (title, date, total_seats, available_seats, booking_ttl)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	err := r.db.Master.QueryRowContext(
		ctx, query, event.Title, event.Date, event.TotalSeats, event.AvailableSeats, event.BookingTTL,
	).Scan(&event.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event.ID, nil
}

// CreateBooking adds a new booking to the database.
func (r *Repository) CreateBooking(ctx context.Context, booking *model.Booking) error {
	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	updateEventQuery := `
		UPDATE events
		SET available_seats = available_seats - 1,
		    updated_at = NOW()
		WHERE id = $1 AND available_seats > 0;
	`

	res, err := tx.ExecContext(ctx, updateEventQuery, booking.EventID)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNoSeatsAvailable
	}

	createBookingQuery := `
		INSERT INTO bookings (event_id, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, status, created_at, updated_at;
	`

	err = tx.QueryRowContext(ctx, createBookingQuery, booking.EventID, booking.UserID, booking.ExpiresAt).
		Scan(&booking.ID, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert booking: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetEventByID retrieves an event by its id.
func (r *Repository) GetEventByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error) {
	query := `
		SELECT id, title, date, total_seats, available_seats, booking_ttl, created_at, updated_at
		FROM events
		WHERE id = $1;
	`

	var event model.Event
	err := r.db.Master.QueryRowContext(
		ctx, query, eventID,
	).Scan(
		&event.ID, &event.Title, &event.Date, &event.TotalSeats, &event.AvailableSeats,
		&event.BookingTTL, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}

		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	return &event, nil
}

// ConfirmBooking sets booking status to confirmed.
func (r *Repository) ConfirmBooking(ctx context.Context, userID, eventID, bookingID uuid.UUID) error {
	query := `
		UPDATE bookings
		SET status = 'confirmed',
		    updated_at = NOW()
		WHERE id = $1 AND event_id = $2 AND user_id = $3 AND status = 'pending'
		RETURNING id;
	`

	var id uuid.UUID
	err := r.db.Master.QueryRowContext(ctx, query, bookingID, eventID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBookingNotFoundOrAlreadyConfirmed
		}

		return fmt.Errorf("failed to update booking: %w", err)
	}

	return nil
}

// CancelBooking sets booking status to cancelled.
func (r *Repository) CancelBooking(ctx context.Context, userID, eventID, bookingID uuid.UUID) error {
	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	cancelQuery := `
		UPDATE bookings
		SET status = 'cancelled',
		    updated_at = NOW()
		WHERE id = $1 AND event_id = $2 AND user_id = $3 AND status = 'pending' AND expires_at < NOW()
		RETURNING id;
    `

	var id uuid.UUID
	err = tx.QueryRowContext(ctx, cancelQuery, bookingID, eventID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBookingNotFoundOrAlreadyCancelled
		}

		return fmt.Errorf("failed to update booking: %w", err)
	}

	updateEventQuery := `
 		UPDATE events
		SET available_seats = available_seats + 1,
		    updated_at = NOW()
 		WHERE id = $1;
	`

	_, err = tx.ExecContext(ctx, updateEventQuery, eventID)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CancelExpiredBookings cancels all expired pending bookings and updates event seats.
func (r *Repository) CancelExpiredBookings(ctx context.Context) (int64, error) {
	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Cancel expired bookings and collect event IDs.
	query := `
		WITH expired AS (
			UPDATE bookings
			SET status = 'cancelled', updated_at = NOW()
			WHERE status = 'pending' AND expires_at < NOW()
			RETURNING event_id
		)
		UPDATE events
		SET available_seats = available_seats + 1,
		    updated_at = NOW()
		FROM expired
		WHERE events.id = expired.event_id
		RETURNING events.id;
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to update events: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		count++
	}

	if err = rows.Err(); err != nil {
		return 0, fmt.Errorf("failed to iterate events: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}
