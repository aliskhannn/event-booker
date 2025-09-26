package scheduler

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/event-booker/internal/model"
)

// userService defines the user service interface used by the CancelExpiredBookingsJob.
type userService interface {
	// GetUserByID returns user info.
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
}

// eventService defines the event-related business logic interface
// that the CancelExpiredBookingsJob depends on.
type eventService interface {
	// GetEventByID returns event info with available seats.
	GetEventByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error)

	// GetExpiredBookings returns all expired bookings (background job).
	GetExpiredBookings(ctx context.Context) ([]*model.Booking, error)

	// CancelBooking cancels a booking (by user or background job).
	CancelBooking(ctx context.Context, userID, eventID, bookingID uuid.UUID) error
}

// notifier defines an interface for sending notifications through a channel.
type notifier interface {
	// Send sends a notification message to the specified recipient.
	Send(to string, msg string) error
}

// CancelExpiredBookingsJob is a background job that cancels expired bookings
// and notifies users via email.
type CancelExpiredBookingsJob struct {
	userService  userService
	eventService eventService
	notifier     notifier
}

// NewCancelExpiredBookingsJob creates a new instance of CancelExpiredBookingsJob
// with the required user service, event service, and notifier.
func NewCancelExpiredBookingsJob(
	userSvc userService,
	eventSvc eventService,
	notifier notifier,
) *CancelExpiredBookingsJob {
	return &CancelExpiredBookingsJob{
		userService:  userSvc,
		eventService: eventSvc,
		notifier:     notifier,
	}
}

// Name returns the name of the job.
func (j *CancelExpiredBookingsJob) Name() string {
	return "CancelExpiredBookingsJob"
}

// Schedule returns the cron schedule for the job.
func (j *CancelExpiredBookingsJob) Schedule() string {
	return "*/30 * * * * *" // runs every 30 seconds
}

// Run executes the job logic: cancel expired bookings and notify users.
func (j *CancelExpiredBookingsJob) Run(ctx context.Context) error {
	// Retrieve all expired bookings.
	booking, err := j.eventService.GetExpiredBookings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired bookings: %w", err)
	}

	for _, b := range booking {
		// Cancel booking.
		if err := j.eventService.CancelBooking(ctx, b.UserID, b.EventID, b.ID); err != nil {
			zlog.Logger.Printf("failed to cancel booking %s: %v", b.ID, err)
			continue
		}

		// Get user by id.
		user, err := j.userService.GetUserByID(ctx, b.UserID)
		if err != nil {
			zlog.Logger.Printf("failed to get user %s: %v", b.UserID, err)
			continue
		}

		// Get event by id.
		event, err := j.eventService.GetEventByID(ctx, b.EventID)
		if err != nil {
			zlog.Logger.Printf("failed to get event %s: %v", b.EventID, err)
			continue
		}

		// Notify user.
		message := fmt.Sprintf("Your booking for the event \"%s\" has been canceled due to expiration.", event.Title)
		if err := j.notifier.Send(user.Email, message); err != nil {
			zlog.Logger.Printf("failed to send notification for booking %s: %v", b.ID, err)
		}
	}

	return nil
}
