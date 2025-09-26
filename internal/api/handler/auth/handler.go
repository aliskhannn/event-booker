package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/event-booker/internal/api/response"
	userrepo "github.com/aliskhannn/event-booker/internal/repository/user"
	userservice "github.com/aliskhannn/event-booker/internal/service/user"
)

// service defines the user service interface used by the auth handler.
type service interface {
	// Register creates a new user with the given email, name, and password.
	Register(ctx context.Context, email, name, password string) (uuid.UUID, error)

	// Login authenticates a user and returns a signed JWT token.
	Login(ctx context.Context, email, password string) (string, error)
}

// Handler provides HTTP handlers for authentication endpoints.
type Handler struct {
	service   service
	validator *validator.Validate
}

// New creates a new authentication handler.
func New(s service, v *validator.Validate) *Handler {
	return &Handler{
		service:   s,
		validator: v,
	}
}

// RegisterRequest represents the JSON request body for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	Name     string `json:"name"`
}

// LoginRequest represents the JSON request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

// Register handles user registration.
// It validates the request body, calls the service layer to create a user,
// and responds with the created user ID.
// Returns 400 for invalid input, 409 if the user already exists,
// and 500 for unexpected errors.
func (h *Handler) Register(c *ginext.Context) {
	var req RegisterRequest

	// Try to parse JSON from the request body into RegisterRequest struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to bind json")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	// Validate the request fields (email, password, etc.).
	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to validate request")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	// Register a new user.
	id, err := h.service.Register(c.Request.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		// If user already exists, return 409 Conflict.
		if errors.Is(err, userservice.ErrUserAlreadyExists) {
			zlog.Logger.Error().Err(err).Msg("user already exists")
			response.Fail(c, http.StatusConflict, fmt.Errorf("user already exists"))
			return
		}

		// Internal Server Error.
		zlog.Logger.Error().Err(err).Msg("failed to register user")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// On success, return 201 Created with the new user ID.
	response.Created(c, map[string]string{
		"id": id.String(),
	})
}

// Login handles user authentication.
// It validates the request body, calls the service layer to authenticate the user,
// and responds with a JWT token on success.
// Returns 400 for invalid input, 401 for invalid credentials,
// 404 if the user does not exist, and 500 for unexpected errors.
func (h *Handler) Login(c *ginext.Context) {
	var req LoginRequest

	// Try to parse JSON from the request body into LoginRequest struct.
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to bind json")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	// Validate the request fields (email and password).
	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to validate request")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	// Authenticate the user and generate a JWT token.
	token, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Invalid credentials: return 401 Unauthorized.
		if errors.Is(err, userservice.ErrInvalidCredentials) {
			zlog.Logger.Error().Err(err).Msg("invalid credentials")
			response.Fail(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
			return
		}

		// User not found: return 404 Not Found.
		if errors.Is(err, userrepo.ErrUserNotFound) {
			zlog.Logger.Error().Err(err).Msg("user not found")
			response.Fail(c, http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}

		// Internal Server Error.
		zlog.Logger.Error().Err(err).Msg("failed to login")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// On success, return 200 OK with the JWT token.
	response.OK(c, map[string]string{
		"token": token,
	})
}
