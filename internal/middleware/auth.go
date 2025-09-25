package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/event-booker/internal/api/response"
)

var (
	ErrNoToken            = errors.New("missing token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidTokenFormat = errors.New("invalid token format")
	ErrExpiredToken       = errors.New("token had expired")
)

// Auth returns a Gin middleware that validates JWT tokens.
// It expects the token in the "Authorization" header in the format "Bearer <token>".
// If the token is missing, malformed, invalid, or expired, it aborts the request with 401 Unauthorized.
// On success, the middleware sets "userID" in the Gin context for downstream handlers.
func Auth(secret string, ttl time.Duration) ginext.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			response.FailAbort(c, http.StatusUnauthorized, ErrNoToken)
			return
		}

		parts := strings.Split(tokenStr, " ") // Bearer <token>
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.FailAbort(c, http.StatusUnauthorized, ErrInvalidTokenFormat)
			return
		}

		userID, err := validateToken(parts[1], secret)
		if err != nil {
			response.FailAbort(c, http.StatusUnauthorized, ErrInvalidToken)
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// validateToken verifies a JWT token and returns the claims.
func validateToken(tokenStr string, secret string) (uuid.UUID, error) {
	// Parse the token.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrExpiredToken
		}

		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}
