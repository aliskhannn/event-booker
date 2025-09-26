package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/aliskhannn/event-booker/internal/config"
	"github.com/aliskhannn/event-booker/internal/model"
	userrepo "github.com/aliskhannn/event-booker/internal/repository/user"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// repository defines the interface for user-related data access.
type repository interface {
	// CreateUser inserts a new user into the storage and returns its ID.
	CreateUser(ctx context.Context, user *model.User) (uuid.UUID, error)

	// GetUserByID retrieves a user by id.
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)

	// GetUserByEmail retrieves a user by their email.
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)

	// CheckUserExistsByEmail checks if a user exists for the given email.
	CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
}

// Service contains business logic for user management such as registration and authentication.
type Service struct {
	repository repository
	cfg        *config.Config
}

// NewService creates a new user service with the provided repository and configuration.
func NewService(r repository, cfg *config.Config) *Service {
	return &Service{
		repository: r,
		cfg:        cfg,
	}
}

// Register creates a new user account with the given email, name, and password.
// It returns the created user's ID or an error if the user already exists or persistence fails.
func (s *Service) Register(ctx context.Context, email, name, password string) (uuid.UUID, error) {
	// Check if user already exists.
	exists, err := s.repository.CheckUserExistsByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("check user exists by email: %w", err)
	}
	if exists {
		return uuid.Nil, ErrUserAlreadyExists
	}

	// Hash password.
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Email:    email,
		Name:     name,
		Password: hashedPassword,
	}

	id, err := s.repository.CreateUser(ctx, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

// Login authenticates a user by email and password and returns a signed JWT token if successful.
// Returns ErrInvalidCredentials if the user does not exist or the password is incorrect.
func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf("get user by email: %w", err)
	}

	// Verify password.
	if err := verifyPassword(password, user.Password); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token.
	token, err := generateToken(user, s.cfg.JWT.Secret, s.cfg.JWT.TTL)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

// GetUserByID returns user info.
func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// hashPassword generates a bcrypt hash for the given password.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hash), err
}

// verifyPassword verifies if the given password matches the stored hash.
func verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateToken creates a signed JWT token containing the user's ID, name, and email.
// The token expires after the configured TTL.
func generateToken(user *model.User, secret string, ttl time.Duration) (string, error) {
	expTime := time.Now().Add(ttl)

	// Create the JWT claims.
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"name":    user.Name,
		"email":   user.Email,
		"exp":     expTime.Unix(),    // expiration time
		"iat":     time.Now().Unix(), // issued at time
	}

	// Create the token with claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key and return.
	return token.SignedString([]byte(secret))
}
