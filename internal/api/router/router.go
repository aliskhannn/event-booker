package router

import (
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/event-booker/internal/api/handler/auth"
	"github.com/aliskhannn/event-booker/internal/api/handler/event"
)

// New creates a new Gin engine and sets up routes for the API.
func New(authHandler *auth.Handler, eventHandler *event.Handler) *ginext.Engine {
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
		// Create a new event
		eventGroup.POST("/", eventHandler.CreateEvent)

		// Retrieve a specific event by ID
		eventGroup.GET("/:eventID", eventHandler.GetEvent)

		// Book a seat for a specific event
		eventGroup.POST("/:eventID/book", eventHandler.BookEvent)

		// Confirm a specific booking
		eventGroup.POST("/:eventID/booking/:bookingID/confirm", eventHandler.ConfirmBooking)

		// Cancel a specific booking
		eventGroup.POST("/:eventID/booking/:bookingID/cancel", eventHandler.CancelBooking)
	}

	return e
}
