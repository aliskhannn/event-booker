# EventBooker

EventBooker is a booking service with deadlines for events like concerts, workshops, courses, and lectures. It allows users to create events, book seats, confirm payments, and automatically cancels unpaid bookings after a specified time interval. This prevents "dead souls" from holding seats, freeing them for other users. The service includes a backend API built with Go, a PostgreSQL database, and a simple React frontend for user interaction.

The backend uses a background scheduler (cron-like) to handle expired bookings. Database operations are transaction-safe to avoid data races.

## Features

- Create events with title, date, total/available seats, and customizable booking TTL.
- Book seats for events (pending status).
- Confirm bookings (payment simulation).
- Cancel bookings (manual or automatic via expiration).
- View event details and available seats.
- Automatic cancellation of expired bookings via a background process.
- User registration and authentication with JWT.
- Email notifications for booking cancellations (using SMTP, e.g., Mailtrap).
- Support for multiple users, with bookings tracked by user ID.
- Simple web UI for creating events, listing events, booking/confirming seats, and observing expiration.

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/aliskhannn/event-booker.git
   cd event-booker
   ```

## Configuration (.env)

Before running the project, copy `.env.example` to `.env` and set your own values:

```
cp .env.example .env
```

Notes:
- SMTP credentials: Create an account on a service like Mailtrap and copy the SMTP details into .env for email notifications.

## Running the Project

Use Makefile for Docker-based setup:

- Build and start all Docker services:
  ```
  make docker-up
  ```

- Stop and remove all Docker services and volumes:
  ```
  make docker-down
  ```

The backend will be available at:
- Backend API → http://localhost:8080/api/

The frontend will be available at:
- Frontend UI → http://localhost:3000

## API Routes

### Auth Routes
- `POST /api/auth/register`: Register a new user. Body: `{ "email": string, "password": string, "name": string }`
- `POST /api/auth/login`: Login and get JWT token. Body: `{ "email": string, "password": string }`

### Event Routes
- `GET /api/events`: List all events (public).
- `GET /api/events/:eventID`: Get event details by ID (public).
- `POST /api/events`: Create a new event (protected). Body: `{ "title": string, "date": string (RFC3339), "total_seats": int, "available_seats": int, "booking_ttl": string (e.g., "10m") }`
- `POST /api/events/:eventID/book`: Book a seat for an event (protected).
- `POST /api/events/:eventID/booking/:bookingID/confirm`: Confirm a booking (protected).
- `POST /api/events/:eventID/booking/:bookingID/cancel`: Cancel a booking (protected).

Protected routes require JWT in `Authorization: Bearer <token>` header.

## Project Structure

```
.
├── cmd/                 # Application entry points
├── config/              # Configuration files
├── internal/            # Internal application packages
│   ├── api/             # HTTP handlers, router, server
│   ├── config/          # Config parsing logic
│   ├── middleware/      # HTTP middlewares
│   ├── model/           # Data models
│   ├── repository/      # Database repositories
│   ├── scheduler/       # cron-like scheduler
│   └── service/         # Business logic
├── migrations/          # Database migrations
├── web/                 # Frontend UI (React + TS + TailwindCSS)
├── Dockerfile           # Backend Dockerfile
├── go.mod
└── go.sum
├── .env.example             # Example environment variables
├── docker-compose.yml       # Multi-service Docker setup
├── Makefile                 # Development commands
└── README.md
```

## Additional Notes

- **Background Scheduler**: Uses a cron-like system to periodically check and cancel expired bookings.
- **Email Notifications**: Implemented for booking cancellations (configurable via SMTP in .env).
- **User Support**: Multiple users can register; bookings are associated with user IDs.
- **Custom TTL**: Each event can have a different booking expiration time.
- **Testing**: Use the UI to create events, book/confirm seats, and observe automatic cancellations after TTL expires.
- **Dependencies**: Backend: Go, Gin, PostgreSQL, Goose for migrations, JWT for auth. Frontend: React, TypeScript, TailwindCSS, Axios.