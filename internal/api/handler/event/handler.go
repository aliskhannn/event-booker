package event

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/event-booker/internal/api/response"
	"github.com/aliskhannn/event-booker/internal/model"
	eventrepo "github.com/aliskhannn/event-booker/internal/repository/event"
	eventservice "github.com/aliskhannn/event-booker/internal/service/event"
)

// service defines the event-related business logic interface
// that the handler depends on.
type service interface {
	// CreateEvent creates new event.
	CreateEvent(ctx context.Context, title string, date time.Time, totalSeats, availableSeats int, bookingTTL time.Duration) (uuid.UUID, error)

	// BookEvent reserves seats for a user at an event.
	BookEvent(ctx context.Context, userID, eventID uuid.UUID) (uuid.UUID, error)

	// GetEventByID returns event info with available seats.
	GetEventByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error)

	// ConfirmBookingPayment confirms the payment of a booking.
	ConfirmBookingPayment(ctx context.Context, bookingID uuid.UUID) error

	// CancelBooking cancels a booking (by user or background job).
	CancelBooking(ctx context.Context, bookingID uuid.UUID) error
}

// Handler provides HTTP endpoints for event management and bookings.
type Handler struct {
	service   service
	validator *validator.Validate
}

// NewHandler creates a new event handler with the provided service and validator.
func NewHandler(s service, v *validator.Validate) *Handler {
	return &Handler{
		service:   s,
		validator: v,
	}
}

// CreateRequest represents the JSON request body for creating an event.
type CreateRequest struct {
	Title          string `json:"title" validate:"required"`
	Date           string `json:"date" validate:"required"`
	TotalSeats     int    `json:"total_seats" validate:"required"`
	AvailableSeats int    `json:"available_seats" validate:"required"`
	BookingTTL     string `json:"booking_ttl"`
}

// CreateEvent handles event creation requests.
// It parses and validates the input, converts date and TTL fields,
// calls the service layer, and returns the created event ID.
func (h *Handler) CreateEvent(c *ginext.Context) {
	var req CreateRequest

	// Try to parse JSON from the request body into CreateRequest struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to bind json")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	// Validate the request fields (title, date, etc.).
	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to validate request")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	// Parse event date.
	eventDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to parse date")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid date"))
		return
	}

	// Parse booking TTL, default to 30 minutes.
	bookingTTL := 30 * time.Minute
	if req.BookingTTL != "" {
		bookingTTL, err = time.ParseDuration(req.BookingTTL)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("failed to parse booking_ttl")
			response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid booking_ttl"))
			return
		}
	}

	// Create a new event.
	id, err := h.service.CreateEvent(
		c.Request.Context(), req.Title, eventDate, req.TotalSeats, req.AvailableSeats, bookingTTL)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to create event")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return created event ID.
	response.Created(c, map[string]string{
		"id": id.String(),
	})
}

// BookEvent handles event booking requests.
// It validates user authorization, parses the event ID,
// and attempts to reserve a seat for the user.
func (h *Handler) BookEvent(c *ginext.Context) {
	userID, err := getUserID(c)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("unauthorized")
		response.Fail(c, http.StatusUnauthorized, err)
		return
	}

	eventID, err := parseUUIDParam(c, "eventID")
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("missing or invalid event id")
		response.Fail(c, http.StatusBadRequest, err)
		return
	}

	// Book a seat.
	id, err := h.service.BookEvent(c.Request.Context(), userID, eventID)
	if err != nil {
		// If  event not found, return 404 Not Found.
		if errors.Is(err, eventservice.ErrEventNotFound) {
			zlog.Logger.Error().Err(err).Msg("event not found")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		// If no seats available, return 409 Conflict.
		if errors.Is(err, eventservice.ErrNoSeatsAvailable) {
			zlog.Logger.Error().Err(err).Msg("no seats available")
			response.Fail(c, http.StatusConflict, err)
			return
		}

		// Internal Server Error.
		zlog.Logger.Error().Err(err).Msg("failed to book event")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return success message.
	response.OK(c, map[string]string{
		"id":      id.String(),
		"message": "booking created",
	})
}

// GetEvent handles requests to fetch a single event by ID.
// It validates the event ID and calls the service to fetch event data.
func (h *Handler) GetEvent(c *ginext.Context) {
	eventID, err := parseUUIDParam(c, "eventID")
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("missing or invalid event id")
		response.Fail(c, http.StatusBadRequest, err)
		return
	}

	// Get an event by its id.
	event, err := h.service.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		// If  event not found, return 404 Not Found.
		if errors.Is(err, eventrepo.ErrEventNotFound) {
			zlog.Logger.Error().Err(err).Msg("event not found")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		// Internal Server Error.
		zlog.Logger.Error().Err(err).Msg("failed to get event")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return event info.
	response.OK(c, map[string]*model.Event{
		"event": event,
	})
}

// ConfirmBooking handles booking confirmation requests.
// It validates user authorization, event ID, and booking ID,
// then calls the service to confirm the booking payment.
func (h *Handler) ConfirmBooking(c *ginext.Context) {
	bookingID, err := parseUUIDParam(c, "bookingID")
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("missing or invalid event id")
		response.Fail(c, http.StatusBadRequest, err)
		return
	}

	// Confirm booking payment.
	err = h.service.ConfirmBookingPayment(c.Request.Context(), bookingID)
	if err != nil {
		// If booking not found or already confirmed, return 404 Not Found.
		if errors.Is(err, eventrepo.ErrBookingNotFoundOrAlreadyConfirmed) {
			zlog.Logger.Error().Err(err).Msg("booking not found or already confirmed")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		// Internal Server Error.
		zlog.Logger.Error().Err(err).Msg("failed to confirm booking payment")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return success.
	response.OK(c, map[string]string{
		"message": "booking confirmed",
	})
}

// CancelBooking handles booking cancellation requests.
// It validates user authorization, event ID, and booking ID,
// then calls the service to cancel the booking.
func (h *Handler) CancelBooking(c *ginext.Context) {
	bookingID, err := parseUUIDParam(c, "bookingID")
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("missing or invalid event id")
		response.Fail(c, http.StatusBadRequest, err)
		return
	}

	// Cancel booking.
	err = h.service.CancelBooking(c.Request.Context(), bookingID)
	if err != nil {
		// If booking not found or already cancelled, return 404 Not Found.
		if errors.Is(err, eventrepo.ErrBookingNotFoundOrAlreadyCancelled) {
			zlog.Logger.Error().Err(err).Msg("booking not found or already cancelled")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		// Internal Server Error.
		zlog.Logger.Error().Err(err).Msg("failed to cancel booking payment")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Return success.
	response.OK(c, map[string]string{
		"message": "booking cancelled",
	})
}

// getUserID extracts the userID from the request context.
// Returns an error if the userID is missing or invalid.
func getUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, fmt.Errorf("userID not found in context")
	}
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid userID in context")
	}
	return userID, nil
}

// ParseUUIDParam parses a UUID from the URL parameters and logs errors if invalid.
// Returns the UUID and an error if parsing fails.
func parseUUIDParam(c *ginext.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		zlog.Logger.Error().Err(err).Interface(param, idStr).Msg("failed to parse UUID")
		return uuid.Nil, fmt.Errorf("invalid %s", param)
	}

	if id == uuid.Nil {
		zlog.Logger.Warn().Interface(param, id).Msg("missing UUID")
		return uuid.Nil, fmt.Errorf("missing %s", param)
	}

	return id, nil
}
