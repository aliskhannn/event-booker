package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/event-booker/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

// Repository provides methods to interact with users table.
type Repository struct {
	db *dbpg.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser adds a new user to the database.
func (r *Repository) CreateUser(ctx context.Context, user *model.User) (uuid.UUID, error) {
	query := `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	err := r.db.QueryRowContext(
		ctx, query, user.Email, user.Password, user.Name,
	).Scan(&user.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

// GetUserByID retrieves a user by id.
func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	query := `
        SELECT id, email, name, created_at
        FROM users
        WHERE id = $1
    `
	var u model.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&u.ID, &u.Email, &u.Name, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return &u, nil
}

// GetUserByEmail retrieves a user by email.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, created_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// CheckUserExistsByEmail checks if a user with the given email already exists in the database.
func (r *Repository) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}
