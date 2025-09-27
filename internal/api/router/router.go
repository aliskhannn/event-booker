package router

import (
	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/event-booker/internal/api/handler/auth"
	"github.com/aliskhannn/event-booker/internal/api/handler/event"
	"github.com/aliskhannn/event-booker/internal/config"
	"github.com/aliskhannn/event-booker/internal/middleware"
)

// New creates a new Gin engine and sets up routes for the API.
func New(authHandler *auth.Handler, eventHandler *event.Handler, cfg *config.Config) *ginext.Engine {
	// Create a new Gin engine using the extended gin wrapper.
	e := ginext.New()

	// Apply global middlewares: CORS (using gin-contrib for reliability), request logging, and panic recovery.
	e.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // Add your frontend ports; use "*" for testing
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))
	e.Use(ginext.Logger())
	e.Use(ginext.Recovery())

	// --- Auth routes ---
	authGroup := e.Group("/api/auth")
	{
		// Register a new user
		authGroup.POST("/register", authHandler.Register)

		// Login user and return JWT token
		authGroup.POST("/login", authHandler.Login)
	}

	// --- Event routes ---
	eventGroup := e.Group("/api/events")
	{
		// Public routes: anyone can view event details
		eventGroup.GET("", eventHandler.GetEvents)

		eventGroup.GET("/:eventID", eventHandler.GetEvent)

		// Protected routes: require auth
		eventGroup.Use(middleware.Auth(cfg.JWT.Secret, cfg.JWT.TTL))
		{
			eventGroup.POST("", eventHandler.CreateEvent)
			eventGroup.POST("/:eventID/book", eventHandler.BookEvent)
			eventGroup.POST("/:eventID/booking/:bookingID/confirm", eventHandler.ConfirmBooking)
			eventGroup.POST("/:eventID/booking/:bookingID/cancel", eventHandler.CancelBooking)
		}
	}

	return e
}
