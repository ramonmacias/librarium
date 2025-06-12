package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"librarium/internal/catalog"
	"librarium/internal/http"
	"librarium/internal/postgres"
	"librarium/internal/user"
)

const (
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

// Application holds all the dependencies needed by the librarium application
// to start, run and close.
type Application struct {
	serverAddress string
	server        *http.Server

	isShuttingDown atomic.Bool

	databaseSource *postgres.DataSource
	db             *sql.DB

	// Repositories
	userRepo    user.Repository
	catalogRepo catalog.Repository

	// Controllers
	authController    *http.AuthController
	catalogController *http.CatalogController
}

// NewLibrariumApplication builds a new librarium application using the provided
// options as setup configuration.
// It returns an error in case of failure.
func NewLibrariumApplication(opts ...Option) (*Application, error) {
	a := &Application{}
	for _, opt := range opts {
		opt(a)
	}

	if err := a.setupInfra(); err != nil {
		return nil, fmt.Errorf("error setting up the application infra %w", err)
	}

	if err := a.setupDomain(); err != nil {
		return nil, fmt.Errorf("error setting up the application domain %w", err)
	}

	if err := a.setupServer(); err != nil {
		return nil, fmt.Errorf("error setting up the application server %w", err)
	}

	return a, nil
}

// Run starts the application, which in this case it starts an http server providing an API.
// It listens
func (a *Application) Run() {
	// Setup signal context
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a.server.ListenAndServe()
	// Wait for signal
	<-rootCtx.Done()
	stop()
	a.isShuttingDown.Store(true)
	log.Println("Received shutdown signal, shutting down.")

	// Give time for readiness check to propagate
	time.Sleep(_readinessDrainDelay)
	log.Println("Readiness check propagated, now waiting for ongoing requests to finish.")

	err := a.server.Shutdown()
	if err != nil {
		log.Println("Failed to wait for ongoing requests to finish, waiting for forced cancellation.")
		time.Sleep(_shutdownHardPeriod)
	}
	err = a.db.Close()
	if err != nil {
		log.Println("Failed to wait for ongoing database requests to finish, waiting for forced cancellation")
		time.Sleep(_shutdownHardPeriod)
	}

	log.Println("Application shut down gracefully.")
}
