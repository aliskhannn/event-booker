package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aliskhannn/event-booker/internal/model"
)

var (
	ErrNoSeatsAvailable = errors.New("no seats available")
)

// repository defines the interface for event booking-related data access.
type repository interface {
	CreateEvent(ctx context.Context, event *model.Event) (uuid.UUID, error)
	CreateBooking(ctx context.Context, booking *model.Booking) error
	GetEventByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error)
	ConfirmBooking(ctx context.Context, bookingID, userID, eventID uuid.UUID) error
	CancelBooking(ctx context.Context, bookingID, userID, eventID uuid.UUID) error
	CancelExpiredBookings(ctx context.Context) (int64, error)
}

// Service contains business logic for event booking management.
type Service struct {
	repository repository
}

// New creates a new event service with the provided repository.
func New(r repository) *Service {
	return &Service{repository: r}
}

// CreateEvent creates new event.
func (s *Service) CreateEvent(
	ctx context.Context,
	title string,
	date time.Time,
	totalSeats, availableSeats int,
	bookingTTL time.Duration,
) (uuid.UUID, error) {
	event := &model.Event{
		Title:          title,
		Date:           date,
		TotalSeats:     totalSeats,
		AvailableSeats: availableSeats,
		BookingTTL:     bookingTTL,
	}

	id, err := s.repository.CreateEvent(ctx, event)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create event: %w", err)
	}

	return id, nil
}

// BookEvent reserves seats for a user at an event.
func (s *Service) BookEvent(ctx context.Context, userID, eventID uuid.UUID) error {
	// Load event to check availability and TTL.
	event, err := s.repository.GetEventByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("get event: %w", err)
	}
	if event.AvailableSeats <= 0 {
		return ErrNoSeatsAvailable
	}

	booking := &model.Booking{
		EventID:   eventID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(event.BookingTTL), // calculate expiration time
	}

	if err := s.repository.CreateBooking(ctx, booking); err != nil {
		return fmt.Errorf("create booking: %w", err)
	}

	return nil
}

// GetEventByID returns event info with available seats.
func (s *Service) GetEventByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error) {
	event, err := s.repository.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	return event, nil
}

// ConfirmBookingPayment confirms the payment of a booking.
func (s *Service) ConfirmBookingPayment(ctx context.Context, bookingID, userID, eventID uuid.UUID) error {
	err := s.repository.ConfirmBooking(ctx, bookingID, userID, eventID)
	if err != nil {
		return fmt.Errorf("confirm booking payment: %w", err)
	}

	return nil
}

// CancelBooking cancels a booking (by user or background job).
func (s *Service) CancelBooking(ctx context.Context, bookingID, userID, eventID uuid.UUID) error {
	err := s.repository.CancelBooking(ctx, bookingID, userID, eventID)
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}

	return nil
}

// CancelExpiredBookings cancels all expired bookings (background job).
func (s *Service) CancelExpiredBookings(ctx context.Context) (int64, error) {
	count, err := s.repository.CancelExpiredBookings(ctx)
	if err != nil {
		return 0, fmt.Errorf("cancel expired bookings: %w", err)
	}

	return count, nil
}
