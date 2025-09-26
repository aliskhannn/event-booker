package router

import (
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

	// Apply global middlewares: CORS, request logging, and panic recovery.
	// TODO: add cors middleware
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
		// Public route: anyone can view event details
		eventGroup.GET("/:eventID", eventHandler.GetEvent)

		// Protected routes: require auth
		eventGroup.Use(middleware.Auth(cfg.JWT.Secret, cfg.JWT.TTL))
		{
			eventGroup.POST("/", eventHandler.CreateEvent)
			eventGroup.POST("/:eventID/book", eventHandler.BookEvent)
			eventGroup.POST("/:eventID/booking/:bookingID/confirm", eventHandler.ConfirmBooking)
			eventGroup.POST("/:eventID/booking/:bookingID/cancel", eventHandler.CancelBooking)
		}
	}

	return e
}
