package main

import (
	"context"
	"errors"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aliskhannn/delayed-notifier/pkg/email"
	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/event-booker/internal/api/handler/auth"
	"github.com/aliskhannn/event-booker/internal/api/handler/event"
	"github.com/aliskhannn/event-booker/internal/api/router"
	"github.com/aliskhannn/event-booker/internal/api/server"
	"github.com/aliskhannn/event-booker/internal/config"
	eventrepo "github.com/aliskhannn/event-booker/internal/repository/event"
	userrepo "github.com/aliskhannn/event-booker/internal/repository/user"
	"github.com/aliskhannn/event-booker/internal/scheduler"
	eventservice "github.com/aliskhannn/event-booker/internal/service/event"
	userservice "github.com/aliskhannn/event-booker/internal/service/user"
)

func main() {
	// Setup context to handle SIGINT and SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize logger, configuration and validator.
	zlog.Init()
	cfg := config.MustLoad()
	val := validator.New()

	// Connect to PostgreSQL master and slave databases.
	opts := &dbpg.Options{
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	slaveDNSs := make([]string, 0, len(cfg.Database.Slaves))

	for _, s := range cfg.Database.Slaves {
		slaveDNSs = append(slaveDNSs, s.DSN())
	}

	db, err := dbpg.New(cfg.Database.Master.DSN(), slaveDNSs, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Initialize email client for notifications.
	smtpPort, err := strconv.Atoi(cfg.Email.SMTPPort)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to parse email smtp port")
	}

	emailClient := email.NewClient(
		cfg.Email.SMTPHost,
		smtpPort,
		cfg.Email.Username,
		cfg.Email.Password,
		cfg.Email.From,
	)

	// Initialize user repository, service, and handler for auth endpoints.
	userRepo := userrepo.NewRepository(db)
	userService := userservice.NewService(userRepo, cfg)
	authHandler := auth.NewHandler(userService, val)

	// Initialize event repository, service, and handler for event endpoints.
	eventRepo := eventrepo.NewRepository(db)
	eventService := eventservice.NewService(eventRepo)
	eventHandler := event.NewHandler(eventService, val)

	// Initialize event repository, service, and handler for event endpoints.
	job := scheduler.NewCancelExpiredBookingsJob(userService, eventService, emailClient)

	// Create a new JobManager and register the job.
	jm := scheduler.NewJobManager(ctx)
	jm.RegisterJob(job)

	// Start the job scheduler in a separate goroutine.
	// The scheduler runs in the background and executes jobs according to their cron schedules.
	go jm.StartScheduler()

	// Initialize API router and HTTP server.
	r := router.New(authHandler, eventHandler)
	s := server.New(cfg.Server.HTTPPort, r)

	// Start HTTP server in a separate goroutine.
	go func() {
		if err := s.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	zlog.Logger.Print("shutdown signal received")

	// Gracefully shutdown server with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zlog.Logger.Print("gracefully shutting down server...\n")
	if err := s.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to shutdown server")
	}
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		zlog.Logger.Info().Msg("timeout exceeded, forcing shutdown")
	}

	zlog.Logger.Print("closing master and slave databases...\n")

	// Close master database connection.
	if err := db.Master.Close(); err != nil {
		zlog.Logger.Printf("failed to close master DB: %v", err)
	}

	// Close slave database connections.
	for i, s := range db.Slaves {
		if err := s.Close(); err != nil {
			zlog.Logger.Printf("failed to close slave DB %d: %v", i, err)
		}
	}
}
